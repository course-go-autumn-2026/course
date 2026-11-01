package finish

import (
	"context"
	"time"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/google/uuid"
)

type tripRepository interface {
	Finish(ctx context.Context, id uuid.UUID, finishedAt time.Time) (trip.Trip, error)
	InsertHistory(ctx context.Context, tripID uuid.UUID, fromStatus *string, toStatus, reason string, changedAt time.Time) error
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
