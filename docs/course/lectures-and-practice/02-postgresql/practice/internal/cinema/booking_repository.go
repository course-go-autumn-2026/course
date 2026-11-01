package cinema

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type BookingRepository struct{}

func NewBookingRepository() *BookingRepository {
	return &BookingRepository{}
}

func (r *BookingRepository) Create(
	ctx context.Context,
	db DBTX,
	screeningID int64,
	userID int64,
) (Booking, error) {
	var booking Booking
	err := db.QueryRow(ctx, `
		INSERT INTO bookings (screening_id, user_id)
		VALUES ($1, $2)
		RETURNING id, screening_id, user_id, created_at
	`, screeningID, userID).Scan(
		&booking.ID,
		&booking.ScreeningID,
		&booking.UserID,
		&booking.CreatedAt,
	)
	if err == nil {
		return booking, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.ConstraintName == "bookings_screening_user_unique" {
		return Booking{}, ErrAlreadyBooked
	}
	return Booking{}, fmt.Errorf("create booking: %w", err)
}
