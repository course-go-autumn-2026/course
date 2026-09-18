package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cinema/internal/cinema"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	demoBooking := flag.Bool("demo-booking", false, "book a seat and exit without starting HTTP")
	screeningID := flag.Int64("screening-id", 0, "screening ID for --demo-booking")
	userID := flag.Int64("user-id", 0, "user ID for --demo-booking")
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	logger := log.New(os.Stdout, "cinema: ", log.LstdFlags|log.Lmicroseconds)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	defer pool.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, 3*time.Second)
	defer cancelPing()
	if err := pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}

	screenings := cinema.NewScreeningRepository(pool, logger)
	bookings := cinema.NewBookingRepository()
	txManager := cinema.NewTransactionManager(pool)
	bookingService := cinema.NewBookingService(txManager, screenings, bookings)

	if *demoBooking {
		if *screeningID < 1 || *userID < 1 {
			return errors.New("--screening-id and --user-id must be positive with --demo-booking")
		}
		booking, err := bookingService.Book(ctx, *screeningID, *userID)
		if err != nil {
			return fmt.Errorf("book: %w", err)
		}
		logger.Printf("booking created: %+v", booking)
		return nil
	}

	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8080"
	}
	handler := cinema.NewHTTPHandler(screenings, bookingService, logger)
	server := &http.Server{
		Addr:              address,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Printf("HTTP server is listening on %s", address)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		return nil
	}
}
