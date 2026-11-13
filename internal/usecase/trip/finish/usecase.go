package finish

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

type Usecase struct {
	trips   tripRepository
	history statusHistoryRepository
	tx      txManager
	metrics metrics
	logger  *slog.Logger
	now     func() time.Time
}

func New(
	trips tripRepository,
	history statusHistoryRepository,
	tx txManager,
	metrics metrics,
	logger *slog.Logger,
	now func() time.Time,
) *Usecase {
	return &Usecase{trips: trips, history: history, tx: tx, metrics: metrics, logger: logger, now: now}
}

func (usecase *Usecase) Finish(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	var value trip.Trip

	err := usecase.tx.Do(ctx, func(txCtx context.Context) error {
		loaded, err := usecase.trips.GetForUpdate(txCtx, id)
		if err != nil {
			return err
		}
		if err := loaded.Finish(usecase.now().UTC()); err != nil {
			return err
		}
		if err := usecase.trips.Update(txCtx, loaded); err != nil {
			return err
		}

		fromStatus := string(trip.StatusActive)
		if err := usecase.history.Add(txCtx, model.StatusChange{
			TripID:     loaded.ID,
			FromStatus: &fromStatus,
			ToStatus:   string(loaded.Status),
			Reason:     "trip completed",
		}); err != nil {
			return err
		}

		value = loaded
		return nil
	})
	if err != nil {
		return trip.Trip{}, err
	}

	usecase.metrics.TripCompleted(ctx)
	usecase.logger.LogAttrs(ctx, slog.LevelInfo, "trip completed", slog.String("trip_id", value.ID.String()))
	return value, nil
}
