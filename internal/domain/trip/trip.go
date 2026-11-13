package trip

import (
	"time"

	"github.com/google/uuid"

	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
)

type Trip struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartPoint     model.Coordinates
	EndPoint       model.Coordinates
	Price          int64
	Status         Status
	StartedAt      time.Time
	FinishedAt     *time.Time
	LastPositionAt *time.Time
}

func New(id uuid.UUID, input model.CreateTripInput, startedAt time.Time) (Trip, error) {
	if input.UserID == uuid.Nil {
		return Trip{}, &ValidationError{Field: "user_id", Message: "must be a non-empty UUID"}
	}
	if input.DriverID == uuid.Nil {
		return Trip{}, &ValidationError{Field: "driver_id", Message: "must be a non-empty UUID"}
	}
	if err := validateCoordinates("start_point", input.StartPoint); err != nil {
		return Trip{}, err
	}
	if err := validateCoordinates("end_point", input.EndPoint); err != nil {
		return Trip{}, err
	}
	if input.Price < 0 {
		return Trip{}, &ValidationError{Field: "price", Message: "cannot be negative"}
	}

	return Trip{
		ID:         id,
		UserID:     input.UserID,
		DriverID:   input.DriverID,
		StartPoint: input.StartPoint,
		EndPoint:   input.EndPoint,
		Price:      input.Price,
		Status:     StatusActive,
		StartedAt:  startedAt,
	}, nil
}

func (value *Trip) Finish(finishedAt time.Time) error {
	if value.Status != StatusActive {
		return ErrTripCompleted
	}

	value.Status = StatusCompleted
	value.FinishedAt = &finishedAt
	return nil
}

func validateCoordinates(field string, value model.Coordinates) error {
	if value.Latitude < -90 || value.Latitude > 90 {
		return &ValidationError{Field: field + ".latitude", Message: "must be between -90 and 90"}
	}
	if value.Longitude < -180 || value.Longitude > 180 {
		return &ValidationError{Field: field + ".longitude", Message: "must be between -180 and 180"}
	}

	return nil
}
