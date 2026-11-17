package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	api "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/generated"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

type createUsecaseFunc func(context.Context, model.CreateTripInput) (trip.Trip, error)

func (function createUsecaseFunc) Create(ctx context.Context, input model.CreateTripInput) (trip.Trip, error) {
	return function(ctx, input)
}

type getUsecaseFunc func(context.Context, uuid.UUID) (trip.Trip, error)

func (function getUsecaseFunc) Get(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	return function(ctx, id)
}

type finishUsecaseFunc func(context.Context, uuid.UUID) (trip.Trip, error)

func (function finishUsecaseFunc) Finish(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	return function(ctx, id)
}

type pingerFunc func(context.Context) error

func (function pingerFunc) Ping(ctx context.Context) error {
	return function(ctx)
}

func TestServerCreateTrip(t *testing.T) {
	t.Parallel()

	wantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	server := NewServer(
		createUsecaseFunc(func(_ context.Context, input model.CreateTripInput) (trip.Trip, error) {
			if input.Price != 1450 || input.UserID == uuid.Nil || input.DriverID == uuid.Nil {
				t.Fatalf("unexpected input: %+v", input)
			}
			return testTrip(wantID), nil
		}),
		getUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, nil }),
		finishUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, nil }),
		pingerFunc(func(context.Context) error { return nil }),
		time.Second,
	)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{
        "user_id":"22222222-2222-4222-8222-222222222222",
        "driver_id":"33333333-3333-4333-8333-333333333333",
        "start_point":{"latitude":59.9398,"longitude":30.3146},
        "end_point":{"latitude":59.929,"longitude":30.3626},
        "price":1450
    }`))
	response := httptest.NewRecorder()

	server.CreateTrip(response, request, api.CreateTripParams{})

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if response.Header().Get("Location") != "/api/v1/trips/"+wantID.String() {
		t.Fatalf("location = %q", response.Header().Get("Location"))
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
	}
}

func TestServerCreateTripRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestServer()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{} {}`))
	response := httptest.NewRecorder()

	server.CreateTrip(response, request, api.CreateTripParams{})

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if response.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
	}
}

func TestServerGetTripMapsNotFound(t *testing.T) {
	t.Parallel()

	server := NewServer(
		createUsecaseFunc(func(context.Context, model.CreateTripInput) (trip.Trip, error) { return trip.Trip{}, nil }),
		getUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, trip.ErrTripNotFound }),
		finishUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, nil }),
		pingerFunc(func(context.Context) error { return nil }),
		time.Second,
	)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/trips/11111111-1111-4111-8111-111111111111", nil)
	response := httptest.NewRecorder()

	server.GetTrip(response, request, uuid.New())

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestServerFinishTripMapsCompleted(t *testing.T) {
	t.Parallel()

	server := NewServer(
		createUsecaseFunc(func(context.Context, model.CreateTripInput) (trip.Trip, error) { return trip.Trip{}, nil }),
		getUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, nil }),
		finishUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, trip.ErrTripCompleted }),
		pingerFunc(func(context.Context) error { return nil }),
		time.Second,
	)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/trips/11111111-1111-4111-8111-111111111111/finish", nil)
	response := httptest.NewRecorder()

	server.FinishTrip(response, request, uuid.New())

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
}

func TestServerReadyReturnsUnavailable(t *testing.T) {
	t.Parallel()

	server := NewServer(
		createUsecaseFunc(func(context.Context, model.CreateTripInput) (trip.Trip, error) { return trip.Trip{}, nil }),
		getUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, nil }),
		finishUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, nil }),
		pingerFunc(func(context.Context) error { return errors.New("database unavailable") }),
		time.Second,
	)
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)
	response := httptest.NewRecorder()

	server.Ready(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func newTestServer() *Server {
	return NewServer(
		createUsecaseFunc(func(context.Context, model.CreateTripInput) (trip.Trip, error) { return trip.Trip{}, nil }),
		getUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, nil }),
		finishUsecaseFunc(func(context.Context, uuid.UUID) (trip.Trip, error) { return trip.Trip{}, nil }),
		pingerFunc(func(context.Context) error { return nil }),
		time.Second,
	)
}

func testTrip(id uuid.UUID) trip.Trip {
	return trip.Trip{
		ID:         id,
		UserID:     uuid.New(),
		DriverID:   uuid.New(),
		StartPoint: model.Coordinates{Latitude: 59.9398, Longitude: 30.3146},
		EndPoint:   model.Coordinates{Latitude: 59.929, Longitude: 30.3626},
		Price:      1450,
		Status:     trip.StatusActive,
		StartedAt:  time.Now().UTC(),
	}
}
