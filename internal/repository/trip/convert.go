package trip

import (
	tripdomain "github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func scanTrip(row pgx.Row) (tripdomain.Trip, error) {
	var trip tripdomain.Trip
	var finishedAt pgtype.Timestamptz
	err := row.Scan(
		&trip.ID, &trip.UserID, &trip.DriverID,
		&trip.StartPoint.Latitude, &trip.StartPoint.Longitude,
		&trip.EndPoint.Latitude, &trip.EndPoint.Longitude,
		&trip.Price, &trip.Status, &trip.StartedAt, &finishedAt,
	)
	if err != nil {
		return tripdomain.Trip{}, err
	}
	if finishedAt.Valid {
		finished := finishedAt.Time
		trip.FinishedAt = &finished
	}
	return trip, nil
}
