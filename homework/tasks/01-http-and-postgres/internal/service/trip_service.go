package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/whxtelyy/course/internal/domain"
)

type TripRepository interface {
	Create(ctx context.Context, trip domain.Trip) (domain.Trip, error)
	AddStatusHistory(ctx context.Context, tripID uuid.UUID, fromStatus *string, toStatus string, reason string) error
	Get(ctx context.Context, id uuid.UUID) (domain.Trip, error)
	Finish(ctx context.Context, id uuid.UUID) (domain.Trip, error)

	ReserveIdempotencyKey(ctx context.Context, key uuid.UUID, requestHash string) (domain.IdempotencyRecord, bool, error)
	SetIdempotencyTrip(ctx context.Context, key uuid.UUID, tripID uuid.UUID) error
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type CreateTripRequest struct {
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartPoint     domain.Coordinates
	EndPoint       domain.Coordinates
	Price          int64
	IdempotencyKey *uuid.UUID
	RequestHash    string
}

type CreateTripResult struct {
	Trip     domain.Trip
	Replayed bool
}

type TripService struct {
	repo TripRepository
	tx   TxManager
}

func NewTripService(repo TripRepository, tx TxManager) *TripService {
	return &TripService{repo: repo, tx: tx}
}

func (s *TripService) Create(ctx context.Context, req CreateTripRequest) (CreateTripResult, error) {
	trip := domain.Trip{
		ID:         uuid.New(),
		UserID:     req.UserID,
		DriverID:   req.DriverID,
		StartPoint: req.StartPoint,
		EndPoint:   req.EndPoint,
		Price:      req.Price,
		Status:     "active",
	}

	var replayed bool

	err := s.tx.Do(ctx, func(txCtx context.Context) error {
		if req.IdempotencyKey != nil {
			record, created, err := s.repo.ReserveIdempotencyKey(
				txCtx,
				*req.IdempotencyKey,
				req.RequestHash,
			)
			if err != nil {
				return err
			}

			if !created {
				if record.RequestHash != req.RequestHash {
					return domain.ErrIdempotencyConflict
				}

				trip, err = s.repo.Get(txCtx, record.TripID)
				if err != nil {
					return err
				}

				replayed = true
				return nil
			}
		}

		var err error

		trip, err = s.repo.Create(txCtx, trip)
		if err != nil {
			return err
		}

		toStatus := "active"
		if err := s.repo.AddStatusHistory(
			txCtx,
			trip.ID,
			nil,
			toStatus,
			"trip created",
		); err != nil {
			return err
		}

		if req.IdempotencyKey != nil {
			if err := s.repo.SetIdempotencyTrip(
				txCtx,
				*req.IdempotencyKey,
				trip.ID,
			); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return CreateTripResult{}, err
	}

	return CreateTripResult{
		Trip:     trip,
		Replayed: replayed,
	}, nil
}

func (s *TripService) Get(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	return s.repo.Get(ctx, id)
}

func (s *TripService) Finish(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	var trip domain.Trip
	returnedErr := s.tx.Do(ctx, func(txCtx context.Context) error {
		var err error
		trip, err = s.repo.Finish(txCtx, id)
		if err != nil {
			return err
		}
		fromStatus := "active"
		if err := s.repo.AddStatusHistory(txCtx, id, &fromStatus, "completed", "trip finished"); err != nil {
			return err
		}
		return nil
	})
	if returnedErr != nil {
		return domain.Trip{}, returnedErr
	}
	return trip, nil
}
