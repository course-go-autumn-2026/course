package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

const scope = "example.com/food-observability"

// POST /orders creates a fixed demo order: no request schema, database or inventory.
func ordersHandler(courierURL string, client *http.Client, logger *slog.Logger) (http.Handler, error) {
	tracer := otel.Tracer(scope)
	meter := otel.Meter(scope)
	attempts, err := meter.Int64Counter("orders.attempts")
	if err != nil {
		return nil, err
	}
	duration, err := meter.Float64Histogram("orders.duration", metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.05, 0.1, 0.3, 0.8, 1, 2, 5))
	if err != nil {
		return nil, err
	}
	for _, outcome := range []string{"success", "error"} {
		attempts.Add(context.Background(), 0, metric.WithAttributes(attribute.String("outcome", outcome)))
	}
	client.Transport = otelhttp.NewTransport(client.Transport)

	create := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		ctx, span := tracer.Start(r.Context(), "orders.create")
		defer span.End()

		err := reserveCourier(ctx, courierURL, client) // Ready-made business operation.
		outcome := "success"
		if err != nil {
			outcome = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, "courier reservation failed")
		}

		result := attribute.String("outcome", outcome)
		span.SetAttributes(result)
		attrs := metric.WithAttributes(result)
		attempts.Add(ctx, 1, attrs)
		duration.Record(ctx, time.Since(started).Seconds(), attrs)
		logger.InfoContext(ctx, "order completed", "outcome", outcome)

		if err != nil {
			http.Error(w, "courier reservation failed", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux := http.NewServeMux()
	mux.Handle("POST /orders", otelhttp.NewHandler(create, "POST /orders"))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	return mux, nil
}

func reserveCourier(ctx context.Context, endpoint string, client *http.Client) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	// Read/close the body to finish the client span and reuse the connection.
	_, err = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("courier returned HTTP %d", response.StatusCode)
	}
	return err
}
