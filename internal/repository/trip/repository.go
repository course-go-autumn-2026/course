package trip

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	domaintrip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/infra/postgres/transaction"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

const driverActiveUniqueIndex = "trips_driver_active_unique_idx"

type Repository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
	builder      sq.StatementBuilderType
}

func New(pool *pgxpool.Pool, queryTimeout time.Duration) *Repository {
	return &Repository{
		pool:         pool,
		queryTimeout: queryTimeout,
		builder:      sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (repository *Repository) Save(ctx context.Context, value domaintrip.Trip) error {
	query, arguments, err := repository.builder.Insert("trips").
		Columns("id", "user_id", "driver_id", "start_latitude", "start_longitude", "end_latitude", "end_longitude", "price", "status", "started_at", "finished_at").
		Values(value.ID, value.UserID, value.DriverID, value.StartPoint.Latitude, value.StartPoint.Longitude, value.EndPoint.Latitude, value.EndPoint.Longitude, value.Price, string(value.Status), value.StartedAt, value.FinishedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	_, err = transaction.FromContext(queryCtx, repository.pool).Exec(queryCtx, query, arguments...)
	if err == nil {
		return nil
	}

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == "23505" && pgError.ConstraintName == driverActiveUniqueIndex {
		return domaintrip.ErrDriverBusy
	}
	return fmt.Errorf("insert trip: %w", err)
}

func (repository *Repository) GetByID(ctx context.Context, id uuid.UUID) (domaintrip.Trip, error) {
	query, arguments, err := repository.selectTrip().Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return domaintrip.Trip{}, fmt.Errorf("build get trip: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	return repository.scanTrip(transaction.FromContext(queryCtx, repository.pool).QueryRow(queryCtx, query, arguments...))
}

func (repository *Repository) GetForUpdate(ctx context.Context, id uuid.UUID) (domaintrip.Trip, error) {
	query, arguments, err := repository.selectTrip().Where(sq.Eq{"id": id}).Suffix("FOR UPDATE").ToSql()
	if err != nil {
		return domaintrip.Trip{}, fmt.Errorf("build get trip for update: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	return repository.scanTrip(transaction.FromContext(queryCtx, repository.pool).QueryRow(queryCtx, query, arguments...))
}

func (repository *Repository) Update(ctx context.Context, value domaintrip.Trip) error {
	query, arguments, err := repository.builder.Update("trips").
		Set("status", string(value.Status)).
		Set("finished_at", value.FinishedAt).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": value.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update trip: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	_, err = transaction.FromContext(queryCtx, repository.pool).Exec(queryCtx, query, arguments...)
	if err != nil {
		return fmt.Errorf("update trip: %w", err)
	}
	return nil
}

func (repository *Repository) Add(ctx context.Context, change model.StatusChange) error {
	query, arguments, err := repository.builder.Insert("trip_status_history").Columns("trip_id", "from_status", "to_status", "reason").Values(change.TripID, change.FromStatus, change.ToStatus, change.Reason).ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip status history: %w", err)
	}
	queryCtx, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()
	if _, err := transaction.FromContext(queryCtx, repository.pool).Exec(queryCtx, query, arguments...); err != nil {
		return fmt.Errorf("insert trip status history: %w", err)
	}
	return nil
}

func (repository *Repository) selectTrip() sq.SelectBuilder {
	return repository.builder.Select("id", "user_id", "driver_id", "start_latitude", "start_longitude", "end_latitude", "end_longitude", "price", "status", "started_at", "finished_at").From("trips")
}

func (repository *Repository) scanTrip(row pgx.Row) (domaintrip.Trip, error) {
	var value domaintrip.Trip
	var status string
	if err := row.Scan(&value.ID, &value.UserID, &value.DriverID, &value.StartPoint.Latitude, &value.StartPoint.Longitude, &value.EndPoint.Latitude, &value.EndPoint.Longitude, &value.Price, &status, &value.StartedAt, &value.FinishedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaintrip.Trip{}, domaintrip.ErrTripNotFound
		}
		return domaintrip.Trip{}, fmt.Errorf("scan trip: %w", err)
	}
	value.Status = domaintrip.Status(status)
	return value, nil
}
