package trip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	tripdomain "github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/model"
)

var tripColumns = []string{
	"id", "user_id", "driver_id",
	"start_latitude", "start_longitude", "end_latitude", "end_longitude",
	"price", "status", "started_at", "finished_at",
}

type Repository struct {
	tx           executorProvider
	queryTimeout time.Duration
}

func NewRepository(tx executorProvider, queryTimeout time.Duration) *Repository {
	return &Repository{tx: tx, queryTimeout: queryTimeout}
}

func (r *Repository) Insert(ctx context.Context, trip tripdomain.Trip) error {
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
			return fmt.Errorf("insert trip: %w", errors.Join(tripdomain.ErrDriverBusy, err))
		}
		return fmt.Errorf("insert trip (SQL %q): %w", query, err)
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
		return fmt.Errorf("insert status history (SQL %q): %w", query, err)
	}
	return nil
}

func (r *Repository) LockIdempotencyKey(ctx context.Context, key uuid.UUID) error {
	query, args, err := squirrel.Select().Column(squirrel.Expr("pg_advisory_xact_lock(hashtextextended(?::text, 0))", key.String())).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build idempotency lock: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	if _, err := r.tx.Executor(queryCtx).Exec(queryCtx, query, args...); err != nil {
		return fmt.Errorf("lock idempotency key (SQL %q): %w", query, err)
	}
	return nil
}

func (r *Repository) GetIdempotency(ctx context.Context, key uuid.UUID) (model.IdempotencyRecord, error) {
	query, args, err := squirrel.Select("request_hash", "trip_snapshot", "expires_at").
		From("trip_idempotency_keys").Where(squirrel.Eq{"key": key}).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return model.IdempotencyRecord{}, fmt.Errorf("build get idempotency key: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	var record model.IdempotencyRecord
	var snapshot []byte
	if err := r.tx.Executor(queryCtx).QueryRow(queryCtx, query, args...).Scan(&record.RequestHash, &snapshot, &record.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.IdempotencyRecord{}, tripdomain.ErrNotFound
		}
		return model.IdempotencyRecord{}, fmt.Errorf("get idempotency key (SQL %q): %w", query, err)
	}
	if err := json.Unmarshal(snapshot, &record.Trip); err != nil {
		return model.IdempotencyRecord{}, fmt.Errorf("decode saved trip: %w", err)
	}
	return record, nil
}

func (r *Repository) DeleteIdempotency(ctx context.Context, key uuid.UUID) error {
	query, args, err := squirrel.Delete("trip_idempotency_keys").
		Where(squirrel.Eq{"key": key}).PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build delete idempotency key: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	if _, err := r.tx.Executor(queryCtx).Exec(queryCtx, query, args...); err != nil {
		return fmt.Errorf("delete idempotency key (SQL %q): %w", query, err)
	}
	return nil
}

func (r *Repository) InsertIdempotency(ctx context.Context, key uuid.UUID, hash [32]byte, trip tripdomain.Trip, expiresAt time.Time) error {
	snapshot, err := json.Marshal(trip)
	if err != nil {
		return fmt.Errorf("encode trip snapshot: %w", err)
	}
	query, args, err := squirrel.Insert("trip_idempotency_keys").
		Columns("key", "request_hash", "trip_id", "trip_snapshot", "expires_at").
		Values(key, hash[:], trip.ID, snapshot, expiresAt).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("build insert idempotency key: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	if _, err := r.tx.Executor(queryCtx).Exec(queryCtx, query, args...); err != nil {
		return fmt.Errorf("insert idempotency key (SQL %q): %w", query, err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (tripdomain.Trip, error) {
	return r.get(ctx, id, false)
}

func (r *Repository) GetForUpdate(ctx context.Context, id uuid.UUID) (tripdomain.Trip, error) {
	return r.get(ctx, id, true)
}

func (r *Repository) get(ctx context.Context, id uuid.UUID, forUpdate bool) (tripdomain.Trip, error) {
	builder := squirrel.Select(tripColumns...).
		From("trips").
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar)
	if forUpdate {
		builder = builder.Suffix("FOR UPDATE")
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return tripdomain.Trip{}, fmt.Errorf("build get trip: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	trip, err := scanTrip(r.tx.Executor(queryCtx).QueryRow(queryCtx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return tripdomain.Trip{}, tripdomain.ErrNotFound
	}
	if err != nil {
		return tripdomain.Trip{}, fmt.Errorf("get trip (SQL %q): %w", query, err)
	}
	return trip, nil
}

func (r *Repository) SaveCompletion(ctx context.Context, trip tripdomain.Trip) error {
	query, args, err := squirrel.Update("trips").
		Set("status", trip.Status).
		Set("finished_at", trip.FinishedAt).
		Set("updated_at", trip.FinishedAt).
		Where(squirrel.Eq{"id": trip.ID, "status": tripdomain.StatusActive}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build save trip completion: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	result, err := r.tx.Executor(queryCtx).Exec(queryCtx, query, args...)
	if err != nil {
		return fmt.Errorf("save trip completion (SQL %q): %w", query, err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("save trip completion (SQL %q): expected one active row, got %d", query, result.RowsAffected())
	}
	return nil
}

func (r *Repository) CountActive(ctx context.Context) (int64, error) {
	queryCtx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()
	var count int64
	const query = "SELECT count(*) FROM trips WHERE status = $1"
	if err := r.tx.Executor(queryCtx).QueryRow(queryCtx, query, tripdomain.StatusActive).Scan(&count); err != nil {
		return 0, fmt.Errorf("count active trips (SQL %q): %w", query, err)
	}
	return count, nil
}
