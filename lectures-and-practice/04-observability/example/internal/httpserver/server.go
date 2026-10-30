// Package httpserver runs the local teaching services and drains their handlers.
package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Serve binds the server and waits for cancellation. Shut down telemetry after it returns.
func Serve(ctx context.Context, address string, handler http.Handler, logger *slog.Logger) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	finished := make(chan error, 1)
	go func() { finished <- server.Serve(listener) }()
	logger.InfoContext(context.Background(), "server started", "address", listener.Addr().String())

	select {
	case err = <-finished:
	case <-ctx.Done():
	}
	// Request contexts are not derived from the signal context: in-flight
	// operations can finish before the SDK providers are shut down.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutdownErr := server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		shutdownErr = errors.Join(shutdownErr, server.Close())
	}
	return errors.Join(err, shutdownErr)
}
