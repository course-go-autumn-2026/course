package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/config"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/container"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run(ctx context.Context) (err error) {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	application, err := container.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := application.Close(context.WithoutCancel(ctx)); closeErr != nil && err == nil {
			err = fmt.Errorf("close application: %w", closeErr)
		}
	}()

	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- application.Server.ListenAndServe()
	}()
	application.Logger.Info("HTTP server started", "address", cfg.HTTPAddr)

	select {
	case serveErr := <-serveErrors:
		if errors.Is(serveErr, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("listen and serve: %w", serveErr)
	case <-ctx.Done():
		application.Logger.Info("shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
		defer cancel()

		if shutdownErr := application.Server.Shutdown(shutdownCtx); shutdownErr != nil {
			return fmt.Errorf("shutdown HTTP server: %w", shutdownErr)
		}
		return nil
	}
}
