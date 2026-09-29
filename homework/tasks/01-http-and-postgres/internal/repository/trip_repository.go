package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whxtelyy/course/internal/domain"
	"github.com/whxtelyy/course/internal/tx"
)

type TripRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewTripRepository(pool *pgxpool.Pool, queryTimeout time.Duration) *TripRepository {
	return &TripRepository{pool: pool, queryTimeout: queryTimeout}
}

func (r *TripRepository) Create(ctx context.Context, trip domain.Trip) (domain.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	executor := tx.ExecutorFromContext(ctx, r.pool)
	query, args, err := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Insert("trips").
		Columns("id", "user_id", "driver_id", "start_latitude", "start_longitude", "end_latitude", "end_longitude", "price", "status", "started_at").
		Values(trip.ID, trip.UserID, trip.DriverID, trip.StartPoint.Latitude, trip.StartPoint.Longitude, trip.EndPoint.Latitude, trip.EndPoint.Longitude, trip.Price, trip.Status, squirrel.Expr("now()")).
		Suffix("RETURNING id, user_id, driver_id, start_latitude, start_longitude, end_latitude, end_longitude, price, status, started_at, finished_at, NULL::timestamptz AS last_position_at").
		ToSql()
	if err != nil {
		return domain.Trip{}, fmt.Errorf("build create trip query: %w", err)
	}

	var result domain.Trip
	err = executor.QueryRow(ctx, query, args...).Scan(
		&result.ID, &result.UserID, &result.DriverID,
		&result.StartPoint.Latitude, &result.StartPoint.Longitude,
		&result.EndPoint.Latitude, &result.EndPoint.Longitude,
		&result.Price, &result.Status, &result.StartedAt, &result.FinishedAt, &result.LastPositionAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "trips_driver_active_uidx" {
			return domain.Trip{}, domain.ErrDriverBusy
		}
		return domain.Trip{}, fmt.Errorf("insert trip: %w", err)
	}
	return result, nil
}

func (r *TripRepository) AddStatusHistory(ctx context.Context, tripID uuid.UUID, fromStatus *string, toStatus string, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	executor := tx.ExecutorFromContext(ctx, r.pool)
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Insert("trip_status_history").
		Columns("trip_id", "from_status", "to_status", "reason").
		Values(tripID, fromStatus, toStatus, reason)
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build status history query: %w", err)
	}
	if _, err := executor.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("insert status history: %w", err)
	}
	return nil
}

func (r *TripRepository) Get(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	executor := tx.ExecutorFromContext(ctx, r.pool)
	query, args, err := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select("id", "user_id", "driver_id", "start_latitude", "start_longitude", "end_latitude", "end_longitude", "price", "status", "started_at", "finished_at", "NULL::timestamptz AS last_position_at").
		From("trips").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return domain.Trip{}, fmt.Errorf("build get trip query: %w", err)
	}

	var trip domain.Trip
	err = executor.QueryRow(ctx, query, args...).Scan(
		&trip.ID, &trip.UserID, &trip.DriverID,
		&trip.StartPoint.Latitude, &trip.StartPoint.Longitude,
		&trip.EndPoint.Latitude, &trip.EndPoint.Longitude,
		&trip.Price, &trip.Status, &trip.StartedAt, &trip.FinishedAt, &trip.LastPositionAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Trip{}, domain.ErrTripNotFound
	}
	if err != nil {
		return domain.Trip{}, fmt.Errorf("get trip: %w", err)
	}
	return trip, nil
}

func (r *TripRepository) Finish(ctx context.Context, id uuid.UUID) (domain.Trip, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	executor := tx.ExecutorFromContext(ctx, r.pool)
	query, args, err := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Update("trips").
		Set("status", "completed").
		Set("finished_at", squirrel.Expr("now()")).
		Set("updated_at", squirrel.Expr("now()")).
		Where(squirrel.Eq{"id": id, "status": "active"}).
		Suffix("RETURNING id, user_id, driver_id, start_latitude, start_longitude, end_latitude, end_longitude, price, status, started_at, finished_at, NULL::timestamptz AS last_position_at").
		ToSql()
	if err != nil {
		return domain.Trip{}, fmt.Errorf("build finish trip query: %w", err)
	}

	var trip domain.Trip
	err = executor.QueryRow(ctx, query, args...).Scan(
		&trip.ID, &trip.UserID, &trip.DriverID,
		&trip.StartPoint.Latitude, &trip.StartPoint.Longitude,
		&trip.EndPoint.Latitude, &trip.EndPoint.Longitude,
		&trip.Price, &trip.Status, &trip.StartedAt, &trip.FinishedAt, &trip.LastPositionAt,
	)
	if err == nil {
		return trip, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Trip{}, fmt.Errorf("finish trip: %w", err)
	}

	_, getErr := r.Get(ctx, id)
	if errors.Is(getErr, domain.ErrTripNotFound) {
		return domain.Trip{}, domain.ErrTripNotFound
	}
	if getErr == nil {
		return domain.Trip{}, domain.ErrTripCompleted
	}
	return domain.Trip{}, getErr
}

func (r *TripRepository) ReserveIdempotencyKey(
	ctx context.Context,
	key uuid.UUID,
	requestHash string,
) (domain.IdempotencyRecord, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	executor := tx.ExecutorFromContext(ctx, r.pool)

	deleteQuery, deleteArgs, err := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Delete("trip_idempotency_keys").
		Where(squirrel.Eq{"idempotency_key": key}).
		Where(squirrel.Expr("expires_at <= now()")).
		ToSql()
	if err != nil {
		return domain.IdempotencyRecord{}, false, fmt.Errorf("build delete expired idempotency key query: %w", err)
	}

	if _, err := executor.Exec(ctx, deleteQuery, deleteArgs...); err != nil {
		return domain.IdempotencyRecord{}, false, fmt.Errorf("delete expired idempotency key: %w", err)
	}

	insertQuery, insertArgs, err := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Insert("trip_idempotency_keys").
		Columns("idempotency_key", "request_hash").
		Values(key, requestHash).
		Suffix("ON CONFLICT (idempotency_key) DO NOTHING RETURNING request_hash").
		ToSql()
	if err != nil {
		return domain.IdempotencyRecord{}, false, fmt.Errorf("build insert idempotency key query: %w", err)
	}

	var storedHash string

	err = executor.QueryRow(ctx, insertQuery, insertArgs...).Scan(&storedHash)
	if err == nil {
		return domain.IdempotencyRecord{}, true, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.IdempotencyRecord{}, false, fmt.Errorf("insert idempotency key: %w", err)
	}

	selectQuery, selectArgs, err := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Select(
			"request_hash",
			"COALESCE(trip_id::text, '')",
		).
		From("trip_idempotency_keys").
		Where(squirrel.Eq{"idempotency_key": key}).
		ToSql()
	if err != nil {
		return domain.IdempotencyRecord{}, false, fmt.Errorf("build get idempotency key query: %w", err)
	}

	var record domain.IdempotencyRecord
	var tripIDText string

	err = executor.QueryRow(ctx, selectQuery, selectArgs...).Scan(
		&record.RequestHash,
		&tripIDText,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.IdempotencyRecord{}, false, fmt.Errorf("idempotency key disappeared")
	}
	if err != nil {
		return domain.IdempotencyRecord{}, false, fmt.Errorf("get idempotency key: %w", err)
	}

	if tripIDText == "" {
		return domain.IdempotencyRecord{}, false, fmt.Errorf("idempotency key has no trip id")
	}

	record.TripID, err = uuid.Parse(tripIDText)
	if err != nil {
		return domain.IdempotencyRecord{}, false, fmt.Errorf("parse idempotency trip id: %w", err)
	}

	return record, false, nil
}

func (r *TripRepository) SetIdempotencyTrip(
	ctx context.Context,
	key uuid.UUID,
	tripID uuid.UUID,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	executor := tx.ExecutorFromContext(ctx, r.pool)

	query, args, err := squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		Update("trip_idempotency_keys").
		Set("trip_id", tripID).
		Where(squirrel.Eq{"idempotency_key": key}).
		Where(squirrel.Eq{"trip_id": nil}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update idempotency key query: %w", err)
	}

	result, err := executor.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update idempotency key: %w", err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("idempotency key update affected %d rows", result.RowsAffected())
	}

	return nil
}
