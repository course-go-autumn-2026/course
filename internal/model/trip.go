package model

import (
	"github.com/google/uuid"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
)

type CreateTripInput struct {
	UserID     uuid.UUID
	DriverID   uuid.UUID
	StartPoint trip.Point
	EndPoint   trip.Point
	Price      int64
}
