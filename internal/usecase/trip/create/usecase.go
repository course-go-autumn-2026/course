package create

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	tripdomain "github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/model"
	"github.com/google/uuid"
)

const idempotencyTTL = 24 * time.Hour

type Usecase struct {
	repo tripRepository
	tx   txManager
}

func New(repo tripRepository, tx txManager) *Usecase {
	return &Usecase{repo: repo, tx: tx}
}

func (s *Usecase) Create(ctx context.Context, input model.CreateTripInput) (tripdomain.Trip, error) {
	trip, _, err := s.CreateWithKey(ctx, input, nil)
	return trip, err
}

func (s *Usecase) CreateWithKey(ctx context.Context, input model.CreateTripInput, key *uuid.UUID) (tripdomain.Trip, bool, error) {
	var requestHash [32]byte
	if key != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return tripdomain.Trip{}, false, fmt.Errorf("encode create request: %w", err)
		}
		requestHash = sha256.Sum256(encoded)
	}

	var trip tripdomain.Trip
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
					return tripdomain.ErrIdempotencyConflict
				}
				trip, replayed = record.Trip, true
				return nil
			case err == nil:
				if err := s.repo.DeleteIdempotency(txCtx, *key); err != nil {
					return err
				}
			case !errors.Is(err, tripdomain.ErrNotFound):
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
		return tripdomain.Trip{}, false, fmt.Errorf("create trip: %w", err)
	}
	return trip, replayed, nil
}

func (s *Usecase) createInTx(ctx context.Context, input model.CreateTripInput) (tripdomain.Trip, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return tripdomain.Trip{}, fmt.Errorf("generate trip ID: %w", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	trip := tripdomain.Trip{
		ID: id, UserID: input.UserID, DriverID: input.DriverID,
		StartPoint: input.StartPoint, EndPoint: input.EndPoint, Price: input.Price,
		Status: tripdomain.StatusActive, StartedAt: now,
	}
	if err := s.repo.Insert(ctx, trip); err != nil {
		return tripdomain.Trip{}, err
	}
	if err := s.repo.InsertHistory(ctx, trip.ID, nil, tripdomain.StatusActive, "Trip created", now); err != nil {
		return tripdomain.Trip{}, err
	}
	return trip, nil
}
