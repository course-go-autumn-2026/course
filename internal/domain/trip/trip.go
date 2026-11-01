package trip

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive    = "active"
	StatusCompleted = "completed"
)

type Point struct {
	Latitude  float64
	Longitude float64
}

type Trip struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	DriverID   uuid.UUID
	StartPoint Point
	EndPoint   Point
	Price      int64
	Status     string
	StartedAt  time.Time
	FinishedAt *time.Time
}

func NewTrip(
	id uuid.UUID,
	userID uuid.UUID,
	driverID uuid.UUID,
	startPoint Point,
	endPoint Point,
	price int64,
	startedAt time.Time,
) (Trip, error) {
	if id == uuid.Nil {
		return Trip{}, &ValidationError{Field: "id", Reason: "must not be zero"}
	}
	if userID == uuid.Nil {
		return Trip{}, &ValidationError{Field: "user_id", Reason: "must not be zero"}
	}
	if driverID == uuid.Nil {
		return Trip{}, &ValidationError{Field: "driver_id", Reason: "must not be zero"}
	}
	if err := validatePoint(startPoint, "start_point"); err != nil {
		return Trip{}, err
	}
	if err := validatePoint(endPoint, "end_point"); err != nil {
		return Trip{}, err
	}
	if price < 0 {
		return Trip{}, &ValidationError{Field: "price", Reason: "must not be negative"}
	}
	if startedAt.IsZero() {
		return Trip{}, &ValidationError{Field: "started_at", Reason: "must not be zero"}
	}

	return Trip{
		ID:         id,
		UserID:     userID,
		DriverID:   driverID,
		StartPoint: startPoint,
		EndPoint:   endPoint,
		Price:      price,
		Status:     StatusActive,
		StartedAt:  startedAt,
		FinishedAt: nil,
	}, nil
}

func validatePoint(point Point, field string) error {
	if !(point.Latitude >= -90 && point.Latitude <= 90) {
		return &ValidationError{Field: field + ".latitude", Reason: "must be between -90 and 90"}
	}
	if !(point.Longitude >= -180 && point.Longitude <= 180) {
		return &ValidationError{Field: field + ".longitude", Reason: "must be between -180 and 180"}
	}
	return nil
}
