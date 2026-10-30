package main

import (
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func couriersHandler(logger *slog.Logger) http.Handler {
	var mode atomic.Value
	mode.Store("normal")
	delays := map[string]time.Duration{
		"normal":  20 * time.Millisecond,
		"slow":    1200 * time.Millisecond,
		"error":   0,
		"timeout": 3 * time.Second, // Longer than the orders client's 2-second timeout.
	}
	mux := http.NewServeMux()
	// Local teaching controls; main.go binds only to loopback. Never expose this
	// unauthenticated endpoint on a shared/public interface.
	mux.HandleFunc("POST /admin/mode/{mode}", func(w http.ResponseWriter, r *http.Request) {
		value := r.PathValue("mode")
		if _, ok := delays[value]; !ok {
			http.Error(w, "mode must be normal, slow, error or timeout", http.StatusBadRequest)
			return
		}
		mode.Store(value)
		w.WriteHeader(http.StatusNoContent)
	})
	reserve := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := mode.Load().(string) // One snapshot per request.
		var failure error
		timer := time.NewTimer(delays[current])
		defer timer.Stop()
		select {
		case <-r.Context().Done():
			failure = r.Context().Err()
		case <-timer.C:
			if current == "error" {
				failure = errors.New("courier unavailable (injected)")
			}
		}
		if failure != nil {
			span := trace.SpanFromContext(r.Context())
			span.RecordError(failure)
			span.SetStatus(codes.Error, "reservation failed")
			logger.ErrorContext(r.Context(), "reservation failed", "mode", current, "error", failure.Error())
			http.Error(w, "reservation failed", http.StatusServiceUnavailable)
			return
		}
		logger.InfoContext(r.Context(), "courier reserved", "mode", current)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.Handle("POST /reserve", otelhttp.NewHandler(reserve, "POST /reserve"))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	return mux
}
