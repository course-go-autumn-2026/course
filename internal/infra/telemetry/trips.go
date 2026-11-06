package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"
)

type activeTripReader interface {
	CountActive(ctx context.Context) (int64, error)
}

type TripEvents struct {
	logger    *slog.Logger
	created   metric.Int64Counter
	completed metric.Int64Counter
}

func NewTripEvents(meter metric.Meter, logger *slog.Logger, reader activeTripReader) (*TripEvents, error) {
	created, err := meter.Int64Counter("tripgo.trips.created", metric.WithUnit("{trip}"))
	if err != nil {
		return nil, fmt.Errorf("create trips counter: %w", err)
	}
	completed, err := meter.Int64Counter("tripgo.trips.completed", metric.WithUnit("{trip}"))
	if err != nil {
		return nil, fmt.Errorf("create completed trips counter: %w", err)
	}
	active, err := meter.Int64ObservableGauge("tripgo.trips.active", metric.WithUnit("{trip}"))
	if err != nil {
		return nil, fmt.Errorf("create active trips gauge: %w", err)
	}
	if _, err := meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		count, err := reader.CountActive(ctx)
		if err != nil {
			logger.WarnContext(ctx, "active trip count failed", "error", err.Error())
			return nil
		}
		observer.ObserveInt64(active, count)
		return nil
	}, active); err != nil {
		return nil, fmt.Errorf("register active trips gauge: %w", err)
	}
	return &TripEvents{logger: logger, created: created, completed: completed}, nil
}

func (e *TripEvents) Created(ctx context.Context, id uuid.UUID) {
	e.created.Add(ctx, 1)
	e.logger.InfoContext(ctx, "trip created", "trip_id", id)
}

func (e *TripEvents) Completed(ctx context.Context, id uuid.UUID) {
	e.completed.Add(ctx, 1)
	e.logger.InfoContext(ctx, "trip completed", "trip_id", id)
}
