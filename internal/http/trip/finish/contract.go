//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package finish_test

package finish

import (
	"context"
	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/google/uuid"
)

type usecase interface {
	Finish(ctx context.Context, id uuid.UUID) (trip.Trip, error)
}
