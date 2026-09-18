package cinema

import (
	"context"
	"errors"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScreeningRepository struct {
	pool   *pgxpool.Pool
	logger *log.Logger
}

func NewScreeningRepository(pool *pgxpool.Pool, logger *log.Logger) *ScreeningRepository {
	return &ScreeningRepository{pool: pool, logger: logger}
}

func (r *ScreeningRepository) Search(
	ctx context.Context,
	filter ScreeningFilter,
) ([]Screening, error) {
	query := sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		Select("id", "film_title", "starts_at", "available_seats").
		From("screenings")

	// TODO(задание 3): добавьте заполненные фильтры, сортировку и Limit.

	querySQL, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build screenings search query: %w", err)
	}
	r.logger.Printf("screenings search SQL: %s; args: %#v", querySQL, args)

	rows, err := r.pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query screenings: %w", err)
	}
	defer rows.Close()

	screenings := make([]Screening, 0)
	for rows.Next() {
		var screening Screening
		if err := rows.Scan(
			&screening.ID,
			&screening.FilmTitle,
			&screening.StartsAt,
			&screening.AvailableSeats,
		); err != nil {
			return nil, fmt.Errorf("scan screening: %w", err)
		}
		screenings = append(screenings, screening)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate screenings: %w", err)
	}

	return screenings, nil
}

func (r *ScreeningRepository) TakeSeat(
	ctx context.Context,
	db DBTX,
	screeningID int64,
) error {
	var remaining int
	err := db.QueryRow(ctx, `
		UPDATE screenings
		SET available_seats = available_seats - 1
		WHERE id = $1 AND available_seats > 0
		RETURNING available_seats
	`, screeningID).Scan(&remaining)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoSeats
	}
	if err != nil {
		return fmt.Errorf("take seat: %w", err)
	}
	return nil
}
