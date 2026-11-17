package trip

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateTripInput struct {
	UserID     uuid.UUID
	DriverID   uuid.UUID
	StartPoint Coordinates
	EndPoint   Coordinates
	Price      int64
}

type Service struct {
	trips   TripRepository
	history StatusHistoryRepository
	tx      TxManager
}

func NewService(
	trips TripRepository,
	history StatusHistoryRepository,
	tx TxManager,
) *Service {
	return &Service{
		trips:   trips,
		history: history,
		tx:      tx,
	}
}

func (service *Service) Create(
	ctx context.Context,
	input CreateTripInput,
) (Trip, error) {
	value := Trip{
		ID:         uuid.New(),
		UserID:     input.UserID,
		DriverID:   input.DriverID,
		StartPoint: input.StartPoint,
		EndPoint:   input.EndPoint,
		Price:      input.Price,
		Status:     StatusActive,
		StartedAt:  time.Now().UTC(),
	}

	err := service.tx.Do(ctx, func(ctx context.Context) error {
		if err := service.trips.Create(ctx, value); err != nil {
			return err
		}

		return service.history.Add(ctx, StatusChange{
			TripID:   value.ID,
			ToStatus: StatusActive,
			Reason:   "trip created",
		})
	})
	if err != nil {
		return Trip{}, err
	}

	return value, nil
}

func (service *Service) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (Trip, error) {
	return service.trips.GetByID(ctx, id)
}

func (service *Service) Finish(
	ctx context.Context,
	id uuid.UUID,
) (Trip, error) {
	var value Trip

	err := service.tx.Do(ctx, func(ctx context.Context) error {
		var err error

		value, err = service.trips.Finish(ctx, id, time.Now().UTC())
		if err != nil {
			return err
		}

		fromStatus := StatusActive
		return service.history.Add(ctx, StatusChange{
			TripID:     value.ID,
			FromStatus: &fromStatus,
			ToStatus:   StatusCompleted,
			Reason:     "trip completed",
		})
	})
	if err != nil {
		return Trip{}, err
	}

	return value, nil
}
