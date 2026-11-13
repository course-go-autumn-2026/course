package get_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/usecase/trip/get"
)

func TestUsecaseGetReturnsRepositoryTrip(t *testing.T) {
	controller := gomock.NewController(t)
	repository := NewMockTripRepository(controller)
	id := uuid.New()
	want := trip.Trip{ID: id, Status: trip.StatusActive, StartedAt: time.Now(), StartPoint: model.Coordinates{}}
	repository.EXPECT().GetByID(gomock.Any(), id).Return(want, nil)

	value, err := get.New(repository).Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if value != want {
		t.Errorf("Get() = %#v, want %#v", value, want)
	}
}

func TestUsecaseGetReturnsNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	repository := NewMockTripRepository(controller)
	id := uuid.New()
	repository.EXPECT().GetByID(gomock.Any(), id).Return(trip.Trip{}, trip.ErrTripNotFound)

	_, err := get.New(repository).Get(context.Background(), id)
	if !errors.Is(err, trip.ErrTripNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, trip.ErrTripNotFound)
	}
}
