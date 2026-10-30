package create

import (
	"context"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package create_test -mock_names usecase=MockUsecase
type usecase interface {
	Create(context.Context, model.CreateTripInput) (trip.Trip, error)
}
