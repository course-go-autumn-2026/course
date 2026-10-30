package postgres

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/trip"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripStatusHistoryRepository struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
	builder      sq.StatementBuilderType
}

func NewTripStatusHistoryRepository(
	pool *pgxpool.Pool,
	queryTimeout time.Duration,
) *TripStatusHistoryRepository {
	return &TripStatusHistoryRepository{
		pool:         pool,
		queryTimeout: queryTimeout,
		builder:      sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (repository *TripStatusHistoryRepository) Add(
	ctx context.Context,
	change trip.StatusChange,
) error {
	var fromStatus any
	if change.FromStatus != nil {
		fromStatus = string(*change.FromStatus)
	}

	query, arguments, err := repository.builder.
		Insert("trip_status_history").
		Columns(
			"trip_id",
			"from_status",
			"to_status",
			"reason",
		).
		Values(
			change.TripID,
			fromStatus,
			string(change.ToStatus),
			change.Reason,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert trip status history: %w", err)
	}

	queryCtx, cancel := context.WithTimeout(ctx, repository.queryTimeout)
	defer cancel()

	_, err = ExecutorFromContext(queryCtx, repository.pool).
		Exec(queryCtx, query, arguments...)
	if err != nil {
		return fmt.Errorf("insert trip status history: %w", err)
	}

	return nil
}
