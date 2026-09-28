package trips

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const idempotencyTTL = 24 * time.Hour

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
	trip, _, err := s.CreateWithKey(ctx, input, nil)
	return trip, err
}

// CreateWithKey returns replayed=true when it returns a saved response for the key.
func (s *Service) CreateWithKey(ctx context.Context, input CreateInput, key *uuid.UUID) (Trip, bool, error) {
	var requestHash [32]byte
	if key != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return Trip{}, false, fmt.Errorf("encode create request: %w", err)
		}
		requestHash = sha256.Sum256(encoded)
	}

	var trip Trip
	var replayed bool
	if err := s.tx.Do(ctx, func(txCtx context.Context) error {
		if key != nil {
			if err := s.repo.LockIdempotencyKey(txCtx, *key); err != nil {
				return err
			}
			record, err := s.repo.GetIdempotency(txCtx, *key)
			switch {
			case err == nil && time.Now().Before(record.ExpiresAt):
				if !bytes.Equal(record.RequestHash, requestHash[:]) {
					return ErrIdempotencyConflict
				}
				trip, replayed = record.Trip, true
				return nil
			case err == nil:
				if err := s.repo.DeleteIdempotency(txCtx, *key); err != nil {
					return err
				}
			case !errors.Is(err, ErrNotFound):
				return err
			}
		}
		var err error
		trip, err = s.createInTx(txCtx, input)
		if err != nil {
			return err
		}
		if key != nil {
			return s.repo.InsertIdempotency(txCtx, *key, requestHash, trip, time.Now().Add(idempotencyTTL))
		}
		return nil
	}); err != nil {
		return Trip{}, false, fmt.Errorf("create trip: %w", err)
	}
	return trip, replayed, nil
}

func (s *Service) createInTx(ctx context.Context, input CreateInput) (Trip, error) {
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
	if err := s.repo.Insert(ctx, trip); err != nil {
		return Trip{}, err
	}
	if err := s.repo.InsertHistory(ctx, trip.ID, nil, StatusActive, "Trip created", now); err != nil {
		return Trip{}, err
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
