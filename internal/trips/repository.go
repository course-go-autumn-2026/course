package trips

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DaleCoopTP/grishin_lab1/internal/database"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var tripColumns = []string{
	"id", "user_id", "driver_id",
	"start_latitude", "start_longitude", "end_latitude", "end_longitude",
	"price", "status", "started_at", "finished_at",
}

type Repository struct {
	tx           *database.Manager
	queryTimeout time.Duration
}

func NewRepository(tx *database.Manager, queryTimeout time.Duration) *Repository {
	return &Repository{tx: tx, queryTimeout: queryTimeout}
}

func (r *Repository) Insert(ctx context.Context, trip Trip) error {
	query, args, err := squirrel.Insert("trips").
		Columns("id", "user_id", "driver_id", "start_latitude", "start_longitude", "end_latitude", "end_longitude", "price", "status", "started_at").
		Values(trip.ID, trip.UserID, trip.DriverID, trip.StartPoint.Latitude, trip.StartPoint.Longitude, trip.EndPoint.Latitude, trip.EndPoint.Longitude, trip.Price, trip.Status, trip.StartedAt).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	if _, err := r.tx.Executor(queryCtx).Exec(queryCtx, query, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "trips_one_active_per_driver_idx" {
			return fmt.Errorf("insert trip: %w", errors.Join(ErrDriverBusy, err))
		}
		return fmt.Errorf("insert trip: %w", err)
	}
	return nil
}

func (r *Repository) InsertHistory(ctx context.Context, tripID uuid.UUID, fromStatus *string, toStatus, reason string, changedAt time.Time) error {
	query, args, err := squirrel.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason", "changed_at").
		Values(tripID, fromStatus, toStatus, reason, changedAt).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert status history: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	if _, err := r.tx.Executor(queryCtx).Exec(queryCtx, query, args...); err != nil {
		return fmt.Errorf("insert status history: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Trip, error) {
	query, args, err := squirrel.Select(tripColumns...).
		From("trips").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return Trip{}, fmt.Errorf("build get trip: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	trip, err := scanTrip(r.tx.Executor(queryCtx).QueryRow(queryCtx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return Trip{}, ErrNotFound
	}
	if err != nil {
		return Trip{}, fmt.Errorf("get trip: %w", err)
	}
	return trip, nil
}

func (r *Repository) Finish(ctx context.Context, id uuid.UUID, finishedAt time.Time) (Trip, error) {
	query, args, err := squirrel.Update("trips").
		Set("status", StatusCompleted).
		Set("finished_at", finishedAt).
		Set("updated_at", finishedAt).
		Where(squirrel.Eq{"id": id, "status": StatusActive}).
		Suffix("RETURNING id, user_id, driver_id, start_latitude, start_longitude, end_latitude, end_longitude, price, status, started_at, finished_at").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return Trip{}, fmt.Errorf("build finish trip: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	trip, err := scanTrip(r.tx.Executor(queryCtx).QueryRow(queryCtx, query, args...))
	if err == nil {
		return trip, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Trip{}, fmt.Errorf("finish trip: %w", err)
	}
	trip, err = r.Get(ctx, id)
	if err != nil {
		return Trip{}, err
	}
	if trip.Status == StatusCompleted {
		return Trip{}, ErrCompleted
	}
	return Trip{}, fmt.Errorf("trip %s is still active after guarded update", id)
}

func scanTrip(row pgx.Row) (Trip, error) {
	var trip Trip
	var finishedAt pgtype.Timestamptz
	err := row.Scan(
		&trip.ID, &trip.UserID, &trip.DriverID,
		&trip.StartPoint.Latitude, &trip.StartPoint.Longitude,
		&trip.EndPoint.Latitude, &trip.EndPoint.Longitude,
		&trip.Price, &trip.Status, &trip.StartedAt, &finishedAt,
	)
	if err != nil {
		return Trip{}, err
	}
	if finishedAt.Valid {
		finished := finishedAt.Time
		trip.FinishedAt = &finished
	}
	return trip, nil
}
