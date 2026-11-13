package finish_test

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
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/usecase/trip/finish"
)

func TestUsecaseFinishCompletesActiveTrip(t *testing.T) {
	controller := gomock.NewController(t)
	repository := NewMockTripRepository(controller)
	history := NewMockStatusHistoryRepository(controller)
	transactions := NewMockTxManager(controller)
	metrics := NewMockMetrics(controller)
	id := uuid.New()
	now := time.Date(2026, time.October, 27, 12, 10, 0, 0, time.UTC)
	value := trip.Trip{ID: id, Status: trip.StatusActive, StartedAt: now.Add(-10 * time.Minute)}

	transactions.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, operation func(context.Context) error) error {
		return operation(ctx)
	})
	repository.EXPECT().GetForUpdate(gomock.Any(), id).Return(value, nil)
	repository.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, updated trip.Trip) error {
		if updated.Status != trip.StatusCompleted {
			t.Errorf("updated status = %q, want %q", updated.Status, trip.StatusCompleted)
		}
		if updated.FinishedAt == nil || !updated.FinishedAt.Equal(now) {
			t.Errorf("updated finished at = %v, want %v", updated.FinishedAt, now)
		}
		return nil
	})
	history.EXPECT().Add(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, change model.StatusChange) error {
		if change.ToStatus != string(trip.StatusCompleted) || change.Reason != "trip completed" {
			t.Errorf("history = %#v, want completed trip history", change)
		}
		return nil
	})
	metrics.EXPECT().TripCompleted(gomock.Any())

	usecase := finish.New(repository, history, transactions, metrics, testLogger(), func() time.Time { return now })
	updated, err := usecase.Finish(context.Background(), id)
	if err != nil {
		t.Fatalf("Finish() error = %v", err)
	}
	if updated.Status != trip.StatusCompleted {
		t.Errorf("Finish() status = %q, want %q", updated.Status, trip.StatusCompleted)
	}
}

func TestUsecaseFinishReturnsCompletedError(t *testing.T) {
	controller := gomock.NewController(t)
	repository := NewMockTripRepository(controller)
	transactions := NewMockTxManager(controller)
	id := uuid.New()

	transactions.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, operation func(context.Context) error) error {
		return operation(ctx)
	})
	repository.EXPECT().GetForUpdate(gomock.Any(), id).Return(trip.Trip{ID: id, Status: trip.StatusCompleted}, nil)

	usecase := finish.New(repository, NewMockStatusHistoryRepository(controller), transactions, NewMockMetrics(controller), testLogger(), time.Now)
	_, err := usecase.Finish(context.Background(), id)
	if !errors.Is(err, trip.ErrTripCompleted) {
		t.Fatalf("Finish() error = %v, want %v", err, trip.ErrTripCompleted)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
