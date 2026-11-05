//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package get_test

package get

import (
	"context"

	"github.com/google/uuid"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
)

type tripRepository interface {
	Get(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}
