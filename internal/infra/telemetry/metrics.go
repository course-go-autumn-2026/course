package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	tripCreated   metric.Int64Counter
	tripCompleted metric.Int64Counter
	httpDuration  metric.Float64Histogram
	httpActive    metric.Int64UpDownCounter
}

func NewMetrics() (*Metrics, error) {
	meter := otel.Meter("trip-service")
	tripCreated, err := meter.Int64Counter("tripgo.trips.created", metric.WithUnit("{trip}"))
	if err != nil {
		return nil, fmt.Errorf("create trip-created metric: %w", err)
	}
	tripCompleted, err := meter.Int64Counter("tripgo.trips.completed", metric.WithUnit("{trip}"))
	if err != nil {
		return nil, fmt.Errorf("create trip-completed metric: %w", err)
	}
	httpDuration, err := meter.Float64Histogram("http.server.request.duration", metric.WithUnit("s"))
	if err != nil {
		return nil, fmt.Errorf("create HTTP duration metric: %w", err)
	}
	httpActive, err := meter.Int64UpDownCounter("http.server.active_requests", metric.WithUnit("{request}"))
	if err != nil {
		return nil, fmt.Errorf("create HTTP active-requests metric: %w", err)
	}

	return &Metrics{
		tripCreated:   tripCreated,
		tripCompleted: tripCompleted,
		httpDuration:  httpDuration,
		httpActive:    httpActive,
	}, nil
}

func (metrics *Metrics) TripCreated(ctx context.Context) {
	metrics.tripCreated.Add(ctx, 1)
}

func (metrics *Metrics) TripCompleted(ctx context.Context) {
	metrics.tripCompleted.Add(ctx, 1)
}

func (metrics *Metrics) RequestStarted(ctx context.Context, method, route string) {
	metrics.httpActive.Add(ctx, 1, metric.WithAttributes(requestAttributes(method, route)...))
}

func (metrics *Metrics) RequestFinished(ctx context.Context, method, route string, status int, seconds float64) {
	attributes := requestAttributes(method, route)
	metrics.httpActive.Add(ctx, -1, metric.WithAttributes(attributes...))
	attributes = append(attributes, attribute.Int("http.response.status_code", status))
	metrics.httpDuration.Record(ctx, seconds, metric.WithAttributes(attributes...))
}

func requestAttributes(method, route string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("http.request.method", method),
		attribute.String("http.route", route),
	}
}
