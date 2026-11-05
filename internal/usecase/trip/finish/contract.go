//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package finish_test

package finish

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
)

type tripRepository interface {
	GetForUpdate(ctx context.Context, id uuid.UUID) (trip.Trip, error)
	SaveCompletion(ctx context.Context, trip trip.Trip) error
	InsertHistory(ctx context.Context, tripID uuid.UUID, fromStatus *string, toStatus, reason string, changedAt time.Time) error
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
