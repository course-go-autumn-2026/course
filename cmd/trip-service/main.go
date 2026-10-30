package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Smoz1e/Bulatov_Ilya_Lab/api"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/config"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/httpapi"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/postgres"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/trip"
	"github.com/go-chi/chi/v5"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return fmt.Errorf("parse LOG_LEVEL: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	tripRepository := postgres.NewTripRepository(
		pool,
		cfg.Database.QueryTimeout,
	)

	historyRepository := postgres.NewTripStatusHistoryRepository(
		pool,
		cfg.Database.QueryTimeout,
	)

	txManager := postgres.NewTransactionManager(pool)

	tripService := trip.NewService(
		tripRepository,
		historyRepository,
		txManager,
	)
	router := chi.NewRouter()

	httpHandler := api.HandlerWithOptions(
		httpapi.NewHandler(
			pool,
			cfg.Database.QueryTimeout,
			tripService,
		),
		api.ChiServerOptions{
			BaseRouter:       router,
			ErrorHandlerFunc: httpapi.HandleOpenAPIError,
		},
	)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpHandler,
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}

	serveErrors := make(chan error, 1)

	go func() {
		serveErrors <- server.ListenAndServe()
	}()

	logger.Info("HTTP server started", "address", cfg.HTTPAddr)

	select {
	case err := <-serveErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("listen and serve: %w", err)

	case <-ctx.Done():
		logger.Info("shutting down HTTP server")

		shutdownCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			cfg.ShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		return nil
	}
}
