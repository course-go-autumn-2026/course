package trip

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TripRepository interface {
	Create(ctx context.Context, value Trip) error
	GetByID(ctx context.Context, id uuid.UUID) (Trip, error)
	Finish(ctx context.Context, id uuid.UUID, finishedAt time.Time) (Trip, error)
}

type StatusChange struct {
	TripID     uuid.UUID
	FromStatus *Status
	ToStatus   Status
	Reason     string
}

type StatusHistoryRepository interface {
	Add(ctx context.Context, change StatusChange) error
}
