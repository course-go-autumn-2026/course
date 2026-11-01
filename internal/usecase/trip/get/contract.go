package get

import (
	"context"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/google/uuid"
)

type tripRepository interface {
	Get(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}
