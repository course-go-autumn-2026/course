package trip_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/google/uuid"
)

func TestNewTrip(t *testing.T) {
	id, userID, driverID := uuid.New(), uuid.New(), uuid.New()
	startedAt := time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC)
	start := trip.Point{Latitude: -90, Longitude: -180}
	end := trip.Point{Latitude: 90, Longitude: 180}

	for _, price := range []int64{0, 1450} {
		got, err := trip.NewTrip(id, userID, driverID, start, end, price, startedAt)
		if err != nil {
			t.Fatalf("price %d: %v", price, err)
		}
		if got.ID != id || got.UserID != userID || got.DriverID != driverID ||
			got.StartPoint != start || got.EndPoint != end || got.Price != price ||
			!got.StartedAt.Equal(startedAt) {
			t.Fatalf("constructor did not preserve input: %+v", got)
		}
		if got.Status != trip.StatusActive || got.FinishedAt != nil {
			t.Fatalf("unexpected initial state: %+v", got)
		}
	}
}

func TestNewTripInvalidInput(t *testing.T) {
	valid := trip.Trip{
		ID: uuid.New(), UserID: uuid.New(), DriverID: uuid.New(),
		StartPoint: trip.Point{Latitude: 59.9, Longitude: 30.3},
		EndPoint:   trip.Point{Latitude: 60, Longitude: 31},
		Price:      1450,
		StartedAt:  time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC),
	}
	cases := []struct {
		name   string
		field  string
		change func(*trip.Trip)
	}{
		{"zero ID", "id", func(v *trip.Trip) { v.ID = uuid.Nil }},
		{"zero user", "user_id", func(v *trip.Trip) { v.UserID = uuid.Nil }},
		{"zero driver", "driver_id", func(v *trip.Trip) { v.DriverID = uuid.Nil }},
		{"start latitude", "start_point.latitude", func(v *trip.Trip) { v.StartPoint.Latitude = -91 }},
		{"start longitude", "start_point.longitude", func(v *trip.Trip) { v.StartPoint.Longitude = 181 }},
		{"end latitude", "end_point.latitude", func(v *trip.Trip) { v.EndPoint.Latitude = 91 }},
		{"end longitude", "end_point.longitude", func(v *trip.Trip) { v.EndPoint.Longitude = -181 }},
		{"NaN latitude", "start_point.latitude", func(v *trip.Trip) { v.StartPoint.Latitude = math.NaN() }},
		{"NaN longitude", "end_point.longitude", func(v *trip.Trip) { v.EndPoint.Longitude = math.NaN() }},
		{"infinite coordinate", "end_point.latitude", func(v *trip.Trip) { v.EndPoint.Latitude = math.Inf(1) }},
		{"negative price", "price", func(v *trip.Trip) { v.Price = -1 }},
		{"zero start time", "started_at", func(v *trip.Trip) { v.StartedAt = time.Time{} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.change(&input)
			got, err := trip.NewTrip(input.ID, input.UserID, input.DriverID,
				input.StartPoint, input.EndPoint, input.Price, input.StartedAt)
			var validationErr *trip.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			if validationErr.Field != tc.field || validationErr.Reason == "" {
				t.Fatalf("unexpected validation details: %+v", validationErr)
			}
			if got != (trip.Trip{}) {
				t.Fatalf("returned a trip despite validation failure: %+v", got)
			}
		})
	}
}
