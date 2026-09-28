package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DaleCoopTP/grishin_lab1/internal/config"
	"github.com/DaleCoopTP/grishin_lab1/internal/database"
	"github.com/DaleCoopTP/grishin_lab1/internal/httpapi"
	"github.com/DaleCoopTP/grishin_lab1/internal/trips"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "configuration:", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})).With("service", "trip-service")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	pool, err := database.Open(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	txManager := database.NewTxManager(pool, cfg.DatabaseQueryTimeout)
	repository := trips.NewRepository(txManager, cfg.DatabaseQueryTimeout)
	service := trips.NewService(repository, txManager)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewHandler(service, pool, cfg.DatabaseQueryTimeout, logger),
		ReadTimeout:       cfg.HTTPReadTimeout,
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
		WriteTimeout:      cfg.HTTPWriteTimeout,
		IdleTimeout:       cfg.HTTPIdleTimeout,
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.HTTPAddr)
	if err != nil {
		pool.Close()
		return fmt.Errorf("listen on %s: %w", cfg.HTTPAddr, err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	logger.Info("HTTP server started", "address", listener.Addr().String())

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			_ = server.Close()
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
			defer cancel()
			if closeErr := closePool(cleanupCtx, pool); closeErr != nil {
				return errors.Join(fmt.Errorf("serve HTTP: %w", err), closeErr)
			}
			return fmt.Errorf("serve HTTP: %w", err)
		}
		pool.Close()
		return nil
	case <-ctx.Done():
		return shutdown(ctx, cfg, logger, server, serveErr, pool)
	}
}

func shutdown(ctx context.Context, cfg config.Config, logger *slog.Logger, server *http.Server, serveErr <-chan error, pool *pgxpool.Pool) error {
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
	defer cancel()
	logger.Info("graceful shutdown started")
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("graceful shutdown exceeded %s: %w", cfg.ShutdownTimeout, err)
	}
	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP during shutdown: %w", err)
		}
	case <-shutdownCtx.Done():
		_ = server.Close()
		return fmt.Errorf("wait for HTTP server: %w", shutdownCtx.Err())
	}
	if err := closePool(shutdownCtx, pool); err != nil {
		return err
	}
	logger.Info("graceful shutdown completed")
	return nil
}

func closePool(ctx context.Context, pool *pgxpool.Pool) error {
	closed := make(chan struct{})
	go func() {
		pool.Close()
		close(closed)
	}()
	select {
	case <-closed:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("close PostgreSQL pool: %w", ctx.Err())
	}
}
