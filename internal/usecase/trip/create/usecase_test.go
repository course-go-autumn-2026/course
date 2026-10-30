package create_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/usecase/trip/create"
)

func TestUsecaseCreatePersistsTripAndHistoryBeforeRecordingMetric(t *testing.T) {
	controller := gomock.NewController(t)
	trips := NewMockTripRepository(controller)
	history := NewMockStatusHistoryRepository(controller)
	transactions := NewMockTxManager(controller)
	metrics := NewMockMetrics(controller)
	now := time.Date(2026, time.October, 27, 12, 0, 0, 0, time.UTC)

	transactions.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, operation func(context.Context) error) error {
			return operation(ctx)
		},
	)
	trips.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, value trip.Trip) error {
			if value.ID == uuid.Nil {
				t.Error("saved trip has empty ID")
			}
			if value.Status != trip.StatusActive {
				t.Errorf("saved trip status = %q, want %q", value.Status, trip.StatusActive)
			}
			if !value.StartedAt.Equal(now) {
				t.Errorf("saved trip started at = %v, want %v", value.StartedAt, now)
			}
			return nil
		},
	)
	history.EXPECT().Add(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, change model.StatusChange) error {
			if change.ToStatus != string(trip.StatusActive) {
				t.Errorf("history status = %q, want %q", change.ToStatus, trip.StatusActive)
			}
			if change.Reason != "trip created" {
				t.Errorf("history reason = %q, want %q", change.Reason, "trip created")
			}
			return nil
		},
	)
	metrics.EXPECT().TripCreated(gomock.Any())

	usecase := create.New(trips, history, transactions, metrics, logger(), func() time.Time { return now })
	value, err := usecase.Create(context.Background(), validInput())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if value.Status != trip.StatusActive {
		t.Errorf("Create() status = %q, want %q", value.Status, trip.StatusActive)
	}
}

func TestUsecaseCreateReturnsDriverBusyWithoutWritingHistory(t *testing.T) {
	controller := gomock.NewController(t)
	trips := NewMockTripRepository(controller)
	transactions := NewMockTxManager(controller)

	transactions.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, operation func(context.Context) error) error {
			return operation(ctx)
		},
	)
	trips.EXPECT().Save(gomock.Any(), gomock.Any()).Return(trip.ErrDriverBusy)

	usecase := create.New(trips, NewMockStatusHistoryRepository(controller), transactions, NewMockMetrics(controller), logger(), time.Now)
	_, err := usecase.Create(context.Background(), validInput())
	if !errors.Is(err, trip.ErrDriverBusy) {
		t.Fatalf("Create() error = %v, want %v", err, trip.ErrDriverBusy)
	}
}

func TestUsecaseCreateReturnsHistoryFailureWithoutRecordingMetric(t *testing.T) {
	controller := gomock.NewController(t)
	trips := NewMockTripRepository(controller)
	history := NewMockStatusHistoryRepository(controller)
	transactions := NewMockTxManager(controller)
	failure := errors.New("history unavailable")

	transactions.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, operation func(context.Context) error) error {
			return operation(ctx)
		},
	)
	trips.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
	history.EXPECT().Add(gomock.Any(), gomock.Any()).Return(failure)

	usecase := create.New(trips, history, transactions, NewMockMetrics(controller), logger(), time.Now)
	_, err := usecase.Create(context.Background(), validInput())
	if !errors.Is(err, failure) {
		t.Fatalf("Create() error = %v, want %v", err, failure)
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

func logger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
