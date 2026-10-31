package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrTripNotFound = errors.New("trip not found")
var ErrDriverBusy = errors.New("driver already has an active trip")
var ErrTripCompleted = errors.New("trip is already completed")

type Trip struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartLatitude  float64
	StartLongitude float64
	EndLatitude    float64
	EndLongitude   float64
	Price          int64
	Status         string
	StartedAt      time.Time
	FinishedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StatusHistory struct {
	ID         int64
	TripID     uuid.UUID
	FromStatus *string
	ToStatus   string
	Reason     *string
	ChangedAt  time.Time
}
