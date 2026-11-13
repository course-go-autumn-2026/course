package trip_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

func TestTripFinishCompletesActiveTrip(t *testing.T) {
	startedAt := time.Date(2026, time.October, 27, 12, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(10 * time.Minute)

	value, err := trip.New(uuid.New(), validInput(), startedAt)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := value.Finish(finishedAt); err != nil {
		t.Fatalf("Finish() error = %v", err)
	}

	if value.Status != trip.StatusCompleted {
		t.Errorf("Status = %q, want %q", value.Status, trip.StatusCompleted)
	}
	if value.FinishedAt == nil || !value.FinishedAt.Equal(finishedAt) {
		t.Errorf("FinishedAt = %v, want %v", value.FinishedAt, finishedAt)
	}
}

func TestTripFinishRejectsCompletedTripWithoutChangingIt(t *testing.T) {
	startedAt := time.Date(2026, time.October, 27, 12, 0, 0, 0, time.UTC)
	firstFinishedAt := startedAt.Add(10 * time.Minute)
	secondFinishedAt := firstFinishedAt.Add(10 * time.Minute)

	value, err := trip.New(uuid.New(), validInput(), startedAt)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := value.Finish(firstFinishedAt); err != nil {
		t.Fatalf("first Finish() error = %v", err)
	}

	err = value.Finish(secondFinishedAt)
	if !errors.Is(err, trip.ErrTripCompleted) {
		t.Fatalf("second Finish() error = %v, want %v", err, trip.ErrTripCompleted)
	}
	if value.Status != trip.StatusCompleted {
		t.Errorf("Status = %q, want %q", value.Status, trip.StatusCompleted)
	}
	if value.FinishedAt == nil || !value.FinishedAt.Equal(firstFinishedAt) {
		t.Errorf("FinishedAt = %v, want %v", value.FinishedAt, firstFinishedAt)
	}
}

func TestNewRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name  string
		input model.CreateTripInput
		field string
	}{
		{name: "empty user id", input: model.CreateTripInput{DriverID: uuid.New()}, field: "user_id"},
		{name: "empty driver id", input: model.CreateTripInput{UserID: uuid.New()}, field: "driver_id"},
		{name: "latitude outside range", input: inputWithStartLatitude(91), field: "start_point.latitude"},
		{name: "negative price", input: inputWithPrice(-1), field: "price"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := trip.New(uuid.New(), testCase.input, time.Now())

			var validationError *trip.ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("New() error = %v, want validation error", err)
			}
			if validationError.Field != testCase.field {
				t.Errorf("ValidationError.Field = %q, want %q", validationError.Field, testCase.field)
			}
		})
	}
}

func validInput() model.CreateTripInput {
	return model.CreateTripInput{
		UserID:   uuid.New(),
		DriverID: uuid.New(),
		StartPoint: model.Coordinates{
			Latitude:  59.9398,
			Longitude: 30.3146,
		},
		EndPoint: model.Coordinates{
			Latitude:  59.929,
			Longitude: 30.3626,
		},
		Price: 1450,
	}
}

func inputWithStartLatitude(latitude float64) model.CreateTripInput {
	input := validInput()
	input.StartPoint.Latitude = latitude
	return input
}

func inputWithPrice(price int64) model.CreateTripInput {
	input := validInput()
	input.Price = price
	return input
}
