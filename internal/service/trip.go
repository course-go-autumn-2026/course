package service

import (
	"context"
	"fmt"
	"time"

	"github.com/TheJCN/Avito-Course/internal/model"
	"github.com/TheJCN/Avito-Course/internal/repository"
	"github.com/google/uuid"
)

type TripService interface {
	Create(ctx context.Context, trip model.Trip) (model.Trip, error)
	FindByID(ctx context.Context, id uuid.UUID) (model.Trip, error)
	Finish(ctx context.Context, id uuid.UUID) (model.Trip, error)
}

type tripService struct {
	trips repository.TripRepository
	tx    repository.TxManager
}

var _ TripService = (*tripService)(nil)

func NewTripService(
	trips repository.TripRepository,
	tx repository.TxManager,
) TripService {
	return &tripService{
		trips: trips,
		tx:    tx,
	}
}

func (s *tripService) Create(ctx context.Context, trip model.Trip) (model.Trip, error) {
	trip.ID = uuid.New()
	trip.Status = "active"
	trip.StartedAt = time.Now().UTC()
	trip.FinishedAt = nil

	err := s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.trips.Create(ctx, trip); err != nil {
			return fmt.Errorf("create trip: %w", err)
		}

		if err := s.addCreatedHistory(ctx, trip.ID); err != nil {
			return err
		}

		var err error
		trip, err = s.trips.FindByID(ctx, trip.ID)
		if err != nil {
			return fmt.Errorf("find created trip: %w", err)
		}

		return nil
	})
	if err != nil {
		return model.Trip{}, err
	}

	return trip, nil
}

func (s *tripService) FindByID(ctx context.Context, id uuid.UUID) (model.Trip, error) {
	trip, err := s.trips.FindByID(ctx, id)
	if err != nil {
		return model.Trip{}, fmt.Errorf("find trip: %w", err)
	}

	return trip, nil
}

func (s *tripService) Finish(ctx context.Context, id uuid.UUID) (model.Trip, error) {
	var trip model.Trip

	err := s.tx.Do(ctx, func(ctx context.Context) error {
		var err error

		trip, err = s.trips.FindByID(ctx, id)
		if err != nil {
			return fmt.Errorf("find trip: %w", err)
		}

		if trip.Status == "completed" {
			return model.ErrTripCompleted
		}

		finishedAt := time.Now().UTC()
		trip.Status = "completed"
		trip.FinishedAt = &finishedAt

		if err := s.trips.Update(ctx, trip); err != nil {
			return fmt.Errorf("finish trip: %w", err)
		}

		if err := s.addFinishedHistory(ctx, trip.ID); err != nil {
			return err
		}

		trip, err = s.trips.FindByID(ctx, id)
		if err != nil {
			return fmt.Errorf("find finished trip: %w", err)
		}

		return nil
	})
	if err != nil {
		return model.Trip{}, err
	}

	return trip, nil
}

func (s *tripService) addCreatedHistory(ctx context.Context, id uuid.UUID) error {
	history := model.StatusHistory{
		TripID:   id,
		ToStatus: "active",
	}

	if err := s.trips.CreateStatusHistory(ctx, history); err != nil {
		return fmt.Errorf("create status history: %w", err)
	}

	return nil
}

func (s *tripService) addFinishedHistory(ctx context.Context, id uuid.UUID) error {
	fromStatus := "active"

	history := model.StatusHistory{
		TripID:     id,
		FromStatus: &fromStatus,
		ToStatus:   "completed",
	}

	if err := s.trips.CreateStatusHistory(ctx, history); err != nil {
		return fmt.Errorf("create status history: %w", err)
	}

	return nil
}
