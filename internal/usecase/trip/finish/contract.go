package finish

import (
	"context"

	"github.com/google/uuid"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package finish_test -mock_names tripRepository=MockTripRepository,statusHistoryRepository=MockStatusHistoryRepository,txManager=MockTxManager,metrics=MockMetrics
type tripRepository interface {
	GetForUpdate(context.Context, uuid.UUID) (trip.Trip, error)
	Update(context.Context, trip.Trip) error
}

type statusHistoryRepository interface {
	Add(context.Context, model.StatusChange) error
}

type txManager interface {
	Do(context.Context, func(context.Context) error) error
}

type metrics interface {
	TripCompleted(context.Context)
}
