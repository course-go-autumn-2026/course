package create

import (
	"context"
	"time"

	tripdomain "github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/model"
	"github.com/google/uuid"
)

type tripRepository interface {
	Insert(ctx context.Context, trip tripdomain.Trip) error
	InsertHistory(ctx context.Context, tripID uuid.UUID, fromStatus *string, toStatus, reason string, changedAt time.Time) error
	LockIdempotencyKey(ctx context.Context, key uuid.UUID) error
	GetIdempotency(ctx context.Context, key uuid.UUID) (model.IdempotencyRecord, error)
	DeleteIdempotency(ctx context.Context, key uuid.UUID) error
	InsertIdempotency(ctx context.Context, key uuid.UUID, hash [32]byte, trip tripdomain.Trip, expiresAt time.Time) error
}

type txManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
