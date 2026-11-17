package create

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
	return &Usecase{
		trips:   trips,
		history: history,
		tx:      tx,
		metrics: metrics,
		logger:  logger,
		now:     now,
	}
}

func (usecase *Usecase) Create(ctx context.Context, input model.CreateTripInput) (trip.Trip, error) {
	value, err := trip.New(uuid.New(), input, usecase.now().UTC())
	if err != nil {
		return trip.Trip{}, err
	}

	err = usecase.tx.Do(ctx, func(txCtx context.Context) error {
		if err := usecase.trips.Save(txCtx, value); err != nil {
			return err
		}

		return usecase.history.Add(txCtx, model.StatusChange{
			TripID:   value.ID,
			ToStatus: string(value.Status),
			Reason:   "trip created",
		})
	})
	if err != nil {
		return trip.Trip{}, err
	}

	usecase.metrics.TripCreated(ctx)
	usecase.logger.LogAttrs(ctx, slog.LevelInfo, "trip created", slog.String("trip_id", value.ID.String()))
	return value, nil
}
