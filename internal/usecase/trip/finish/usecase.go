package finish

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	tripdomain "github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
)

type Usecase struct {
	repo   tripRepository
	tx     txManager
	events tripEvents
}

func New(repo tripRepository, tx txManager, events tripEvents) *Usecase {
	return &Usecase{repo: repo, tx: tx, events: events}
}

func (s *Usecase) Finish(ctx context.Context, id uuid.UUID) (tripdomain.Trip, error) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	var trip tripdomain.Trip
	if err := s.tx.Do(ctx, func(txCtx context.Context) error {
		var err error
		trip, err = s.repo.GetForUpdate(txCtx, id)
		if err != nil {
			return err
		}
		fromStatus := trip.Status
		if err := trip.Finish(now); err != nil {
			return err
		}
		if err := s.repo.SaveCompletion(txCtx, trip); err != nil {
			return err
		}
		return s.repo.InsertHistory(txCtx, id, &fromStatus, trip.Status, "Trip finished", now)
	}); err != nil {
		return tripdomain.Trip{}, fmt.Errorf("finish trip: %w", err)
	}
	if s.events != nil {
		s.events.Completed(ctx, trip.ID)
	}
	return trip, nil
}
