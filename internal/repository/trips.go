package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/TheJCN/Avito-Course/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepository interface {
	Create(ctx context.Context, trip model.Trip) error
	FindByID(ctx context.Context, id uuid.UUID) (model.Trip, error)
	FindAll(ctx context.Context) ([]model.Trip, error)
	Update(ctx context.Context, trip model.Trip) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateStatusHistory(ctx context.Context, history model.StatusHistory) error
}

type executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresTripRepository struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

var _ TripRepository = (*PostgresTripRepository)(nil)

func NewTripRepository(pool *pgxpool.Pool, timeout time.Duration) *PostgresTripRepository {
	return &PostgresTripRepository{pool: pool, timeout: timeout}
}

func (r *PostgresTripRepository) executor(ctx context.Context) executor {
	tx := transactionFromContext(ctx)
	if tx != nil {
		return tx
	}
	return r.pool
}

func (r *PostgresTripRepository) Create(ctx context.Context, trip model.Trip) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query, args, err := sq.Insert("trips").
		Columns("id", "user_id", "driver_id", "start_latitude", "start_longitude", "end_latitude", "end_longitude", "price", "status", "started_at", "finished_at").
		Values(trip.ID, trip.UserID, trip.DriverID, trip.StartLatitude, trip.StartLongitude, trip.EndLatitude, trip.EndLongitude, trip.Price, trip.Status, trip.StartedAt, trip.FinishedAt).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip query: %w", err)
	}

	_, err = r.executor(ctx).Exec(ctx, query, args...)
	if err != nil {
		return tripWriteError("create trip", err)
	}
	return nil
}

func tripSelect() sq.SelectBuilder {
	return sq.Select("id", "user_id", "driver_id", "start_latitude", "start_longitude", "end_latitude", "end_longitude", "price", "status", "started_at", "finished_at", "created_at", "updated_at").
		From("trips").PlaceholderFormat(sq.Dollar)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTrip(row scanner) (model.Trip, error) {
	var trip model.Trip
	err := row.Scan(&trip.ID, &trip.UserID, &trip.DriverID, &trip.StartLatitude, &trip.StartLongitude, &trip.EndLatitude, &trip.EndLongitude, &trip.Price, &trip.Status, &trip.StartedAt, &trip.FinishedAt, &trip.CreatedAt, &trip.UpdatedAt)
	return trip, err
}

func (r *PostgresTripRepository) FindByID(ctx context.Context, id uuid.UUID) (model.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query, args, err := tripSelect().Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return model.Trip{}, fmt.Errorf("build find trip query: %w", err)
	}
	trip, err := scanTrip(r.executor(ctx).QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Trip{}, model.ErrTripNotFound
	}
	if err != nil {
		return model.Trip{}, fmt.Errorf("find trip: %w", err)
	}
	return trip, nil
}

func (r *PostgresTripRepository) FindAll(ctx context.Context) ([]model.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query, args, err := tripSelect().OrderBy("started_at", "id").ToSql()
	if err != nil {
		return nil, fmt.Errorf("build find all trips query: %w", err)
	}
	rows, err := r.executor(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find all trips: %w", err)
	}
	defer rows.Close()

	trips := make([]model.Trip, 0)
	for rows.Next() {
		trip, err := scanTrip(rows)
		if err != nil {
			return nil, fmt.Errorf("read trip: %w", err)
		}
		trips = append(trips, trip)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read trips: %w", err)
	}
	return trips, nil
}

func (r *PostgresTripRepository) Update(ctx context.Context, trip model.Trip) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query, args, err := sq.Update("trips").
		Set("user_id", trip.UserID).Set("driver_id", trip.DriverID).
		Set("start_latitude", trip.StartLatitude).Set("start_longitude", trip.StartLongitude).
		Set("end_latitude", trip.EndLatitude).Set("end_longitude", trip.EndLongitude).
		Set("price", trip.Price).Set("status", trip.Status).Set("finished_at", trip.FinishedAt).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"id": trip.ID, "status": "active"}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build update trip query: %w", err)
	}
	result, err := r.executor(ctx).Exec(ctx, query, args...)
	if err != nil {
		return tripWriteError("update trip", err)
	}
	if result.RowsAffected() == 0 {
		return r.updateError(ctx, trip.ID)
	}
	return nil
}

func (r *PostgresTripRepository) updateError(ctx context.Context, id uuid.UUID) error {
	trip, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if trip.Status == "completed" {
		return model.ErrTripCompleted
	}
	return fmt.Errorf("update trip: no rows updated")
}

func (r *PostgresTripRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query, args, err := sq.Delete("trips").Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build delete trip query: %w", err)
	}
	result, err := r.executor(ctx).Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete trip: %w", err)
	}
	if result.RowsAffected() == 0 {
		return model.ErrTripNotFound
	}
	return nil
}

func (r *PostgresTripRepository) CreateStatusHistory(ctx context.Context, history model.StatusHistory) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query, args, err := sq.Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason").
		Values(history.TripID, history.FromStatus, history.ToStatus, history.Reason).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build insert history query: %w", err)
	}
	_, err = r.executor(ctx).Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("create status history: %w", err)
	}
	return nil
}

func tripWriteError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "trips_active_driver_idx" {
		return fmt.Errorf("%s: %w", operation, model.ErrDriverBusy)
	}
	return fmt.Errorf("%s: %w", operation, err)
}
