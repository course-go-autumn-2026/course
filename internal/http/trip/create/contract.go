//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package create_test

package create

import (
	"context"
	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/model"
	"github.com/google/uuid"
)

type usecase interface {
	CreateWithKey(ctx context.Context, input model.CreateTripInput, key *uuid.UUID) (trip.Trip, bool, error)
}
