package create

import (
	"context"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package create_test -mock_names tripRepository=MockTripRepository,statusHistoryRepository=MockStatusHistoryRepository,txManager=MockTxManager,metrics=MockMetrics
type tripRepository interface {
	Save(context.Context, trip.Trip) error
}

type statusHistoryRepository interface {
	Add(context.Context, model.StatusChange) error
}

type txManager interface {
	Do(context.Context, func(context.Context) error) error
}

type metrics interface {
	TripCreated(context.Context)
}
