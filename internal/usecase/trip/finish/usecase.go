package finish

import (
	"context"
	"fmt"
	"time"

	tripdomain "github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/google/uuid"
)

type Usecase struct {
	repo tripRepository
	tx   txManager
}

func New(repo tripRepository, tx txManager) *Usecase {
	return &Usecase{repo: repo, tx: tx}
}

func (s *Usecase) Finish(ctx context.Context, id uuid.UUID) (tripdomain.Trip, error) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	var trip tripdomain.Trip
	if err := s.tx.Do(ctx, func(txCtx context.Context) error {
		var err error
		trip, err = s.repo.Finish(txCtx, id, now)
		if err != nil {
			return err
		}
		fromStatus := tripdomain.StatusActive
		return s.repo.InsertHistory(txCtx, id, &fromStatus, tripdomain.StatusCompleted, "Trip finished", now)
	}); err != nil {
		return tripdomain.Trip{}, fmt.Errorf("finish trip: %w", err)
	}
	return trip, nil
}
