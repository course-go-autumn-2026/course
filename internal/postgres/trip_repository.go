package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	sq "github.com/Masterminds/squirrel"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/trip"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const driverActiveUniqueIndex = "trips_driver_active_unique_idx"

type TripRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
	builder      sq.StatementBuilderType
}

func NewTripRepository(
	pool *pgxpool.Pool,
	queryTimeout time.Duration,
) *TripRepository {
	return &TripRepository{
		pool:         pool,
		queryTimeout: queryTimeout,
		builder:      sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (repository *TripRepository) Create(
	ctx context.Context,
	value trip.Trip,
) error {
	query, arguments, err := repository.builder.
		Insert("trips").
		Columns(
			"id",
			"user_id",
			"driver_id",
			"start_latitude",
			"start_longitude",
			"end_latitude",
			"end_longitude",
			"price",
			"status",
			"started_at",
			"finished_at",
		).
		Values(
			value.ID,
			value.UserID,
			value.DriverID,
			value.StartPoint.Latitude,
			value.StartPoint.Longitude,
			value.EndPoint.Latitude,
			value.EndPoint.Longitude,
			value.Price,
			string(value.Status),
			value.StartedAt,
			value.FinishedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()

	_, err = ExecutorFromContext(queryCtx, repository.pool).
		Exec(queryCtx, query, arguments...)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == driverActiveUniqueIndex {
		return trip.ErrDriverBusy
	}

	return fmt.Errorf("insert trip: %w", err)
}

func (repository *TripRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (trip.Trip, error) {
	query, arguments, err := repository.builder.
		Select(
			"id",
			"user_id",
			"driver_id",
			"start_latitude",
			"start_longitude",
			"end_latitude",
			"end_longitude",
			"price",
			"status",
			"started_at",
			"finished_at",
		).
		From("trips").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return trip.Trip{}, fmt.Errorf("build get trip: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()

	var value trip.Trip
	var status string

	err = ExecutorFromContext(queryCtx, repository.pool).
		QueryRow(queryCtx, query, arguments...).
		Scan(
			&value.ID,
			&value.UserID,
			&value.DriverID,
			&value.StartPoint.Latitude,
			&value.StartPoint.Longitude,
			&value.EndPoint.Latitude,
			&value.EndPoint.Longitude,
			&value.Price,
			&status,
			&value.StartedAt,
			&value.FinishedAt,
		)
	if errors.Is(err, pgx.ErrNoRows) {
		return trip.Trip{}, trip.ErrTripNotFound
	}
	if err != nil {
		return trip.Trip{}, fmt.Errorf("get trip: %w", err)
	}

	value.Status = trip.Status(status)
	return value, nil
}

func (repository *TripRepository) Finish(
	ctx context.Context,
	id uuid.UUID,
	finishedAt time.Time,
) (trip.Trip, error) {
	query, arguments, err := repository.builder.
		Update("trips").
		Set("status", string(trip.StatusCompleted)).
		Set("finished_at", finishedAt).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"status": string(trip.StatusActive)}).
		Suffix(`
			RETURNING
				id,
				user_id,
				driver_id,
				start_latitude,
				start_longitude,
				end_latitude,
				end_longitude,
				price,
				status,
				started_at,
				finished_at
		`).
		ToSql()
	if err != nil {
		return trip.Trip{}, fmt.Errorf("build finish trip: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()

	var value trip.Trip
	var status string

	err = ExecutorFromContext(queryCtx, repository.pool).
		QueryRow(queryCtx, query, arguments...).
		Scan(
			&value.ID,
			&value.UserID,
			&value.DriverID,
			&value.StartPoint.Latitude,
			&value.StartPoint.Longitude,
			&value.EndPoint.Latitude,
			&value.EndPoint.Longitude,
			&value.Price,
			&status,
			&value.StartedAt,
			&value.FinishedAt,
		)
	if errors.Is(err, pgx.ErrNoRows) {
		_, getErr := repository.GetByID(ctx, id)
		if errors.Is(getErr, trip.ErrTripNotFound) {
			return trip.Trip{}, trip.ErrTripNotFound
		}
		if getErr != nil {
			return trip.Trip{}, getErr
		}

		return trip.Trip{}, trip.ErrTripCompleted
	}
	if err != nil {
		return trip.Trip{}, fmt.Errorf("finish trip: %w", err)
	}

	value.Status = trip.Status(status)
	return value, nil
}
