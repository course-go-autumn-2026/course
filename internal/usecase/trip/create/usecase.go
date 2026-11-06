package create

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	tripdomain "github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/model"
)

const idempotencyTTL = 24 * time.Hour

type Usecase struct {
	repo   tripRepository
	tx     txManager
	events tripEvents
}

func New(repo tripRepository, tx txManager, events tripEvents) *Usecase {
	return &Usecase{repo: repo, tx: tx, events: events}
}

func (s *Usecase) Create(ctx context.Context, input model.CreateTripInput) (tripdomain.Trip, error) {
	trip, _, err := s.CreateWithKey(ctx, input, nil)
	return trip, err
}

func (s *Usecase) CreateWithKey(ctx context.Context, input model.CreateTripInput, key *uuid.UUID) (tripdomain.Trip, bool, error) {
	if err := tripdomain.ValidateCreate(input.UserID, input.DriverID, input.StartPoint, input.EndPoint, input.Price); err != nil {
		return tripdomain.Trip{}, false, err
	}

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
	if s.events != nil && !replayed {
		s.events.Created(ctx, trip.ID)
	}
	return trip, replayed, nil
}

func (s *Usecase) createInTx(ctx context.Context, input model.CreateTripInput) (tripdomain.Trip, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return tripdomain.Trip{}, fmt.Errorf("generate trip ID: %w", err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	trip, err := tripdomain.NewTrip(id, input.UserID, input.DriverID, input.StartPoint, input.EndPoint, input.Price, now)
	if err != nil {
		return tripdomain.Trip{}, err
	}
	if err := s.repo.Insert(ctx, trip); err != nil {
		return tripdomain.Trip{}, err
	}
	if err := s.repo.InsertHistory(ctx, trip.ID, nil, trip.Status, "Trip created", trip.StartedAt); err != nil {
		return tripdomain.Trip{}, err
	}
	return trip, nil
}
