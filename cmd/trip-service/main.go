package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DaleCoopTP/grishin_lab1/internal/config"
	"github.com/DaleCoopTP/grishin_lab1/internal/container"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "configuration:", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	app, err := container.New(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "initialize application:", err)
		os.Exit(1)
	}
	if err := run(ctx, cfg, app); err != nil {
		app.Logger.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config, app *container.Container) error {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", app.Server.Addr)
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
		defer cancel()
		return errors.Join(fmt.Errorf("listen on %s: %w", app.Server.Addr, err), app.Close(cleanupCtx))
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- app.Server.Serve(listener) }()
	app.Logger.Info("HTTP server started", "address", listener.Addr().String())

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		} else {
			_ = app.Server.Close()
			err = fmt.Errorf("serve HTTP: %w", err)
		}
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
		defer cancel()
		return errors.Join(err, app.Close(cleanupCtx))
	case <-ctx.Done():
		return shutdown(ctx, cfg, app, serveErr)
	}
}

func shutdown(ctx context.Context, cfg config.Config, app *container.Container, serveErr <-chan error) error {
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
	defer cancel()
	app.Logger.Info("graceful shutdown started")

	var shutdownErr error
	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		_ = app.Server.Close()
		shutdownErr = fmt.Errorf("graceful shutdown exceeded %s: %w", cfg.ShutdownTimeout, err)
	} else {
		select {
		case err := <-serveErr:
			if !errors.Is(err, http.ErrServerClosed) {
				shutdownErr = fmt.Errorf("serve HTTP during shutdown: %w", err)
			}
		case <-shutdownCtx.Done():
			_ = app.Server.Close()
			shutdownErr = fmt.Errorf("wait for HTTP server: %w", shutdownCtx.Err())
		}
	}
	if err := errors.Join(shutdownErr, app.Close(shutdownCtx)); err != nil {
		return err
	}
	app.Logger.Info("graceful shutdown completed")
	return nil
}
