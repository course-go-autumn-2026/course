package trips

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	repo *Repository
	tx   TxManager
}

func NewService(repo *Repository, tx TxManager) *Service {
	return &Service{repo: repo, tx: tx}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Trip, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return Trip{}, fmt.Errorf("generate trip ID: %w", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	trip := Trip{
		ID: id, UserID: input.UserID, DriverID: input.DriverID,
		StartPoint: input.StartPoint, EndPoint: input.EndPoint, Price: input.Price,
		Status: StatusActive, StartedAt: now,
	}
	if err := s.tx.Do(ctx, func(txCtx context.Context) error {
		if err := s.repo.Insert(txCtx, trip); err != nil {
			return err
		}
		return s.repo.InsertHistory(txCtx, trip.ID, nil, StatusActive, "Trip created", now)
	}); err != nil {
		return Trip{}, fmt.Errorf("create trip: %w", err)
	}
	return trip, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Trip, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Finish(ctx context.Context, id uuid.UUID) (Trip, error) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	var trip Trip
	if err := s.tx.Do(ctx, func(txCtx context.Context) error {
		var err error
		trip, err = s.repo.Finish(txCtx, id, now)
		if err != nil {
			return err
		}
		fromStatus := StatusActive
		return s.repo.InsertHistory(txCtx, id, &fromStatus, StatusCompleted, "Trip finished", now)
	}); err != nil {
		return Trip{}, fmt.Errorf("finish trip: %w", err)
	}
	return trip, nil
}
