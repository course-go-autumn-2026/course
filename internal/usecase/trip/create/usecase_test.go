package create_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/model"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/create"
)

func TestCreate(t *testing.T) {
	storageErr := errors.New("storage failure")
	commitErr := errors.New("commit failure")
	cases := []struct {
		name                             string
		insertErr, historyErr, commitErr error
		wantErr                          error
	}{
		{name: "success"},
		{name: "busy driver", insertErr: trip.ErrDriverBusy, wantErr: trip.ErrDriverBusy},
		{name: "insert failure", insertErr: storageErr, wantErr: storageErr},
		{name: "history failure", historyErr: storageErr, wantErr: storageErr},
		{name: "commit failure", commitErr: commitErr, wantErr: commitErr},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo, tx := NewMocktripRepository(ctrl), NewMocktxManager(ctrl)
			ctx := t.Context()
			txCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			tx.EXPECT().Do(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				if err := fn(txCtx); err != nil {
					return err
				}
				return tc.commitErr
			})
			input := validInput()
			var inserted trip.Trip
			insert := repo.EXPECT().Insert(txCtx, gomock.Any()).DoAndReturn(func(_ context.Context, value trip.Trip) error {
				inserted = value
				if value.ID == uuid.Nil || value.StartedAt.IsZero() || value.FinishedAt != nil || value.Status != trip.StatusActive ||
					value.UserID != input.UserID || value.DriverID != input.DriverID || value.Price != input.Price ||
					value.StartPoint != input.StartPoint || value.EndPoint != input.EndPoint {
					t.Fatalf("unexpected inserted trip: %+v", value)
				}
				return tc.insertErr
			})
			if tc.insertErr == nil {
				repo.EXPECT().InsertHistory(txCtx, gomock.Any(), (*string)(nil), trip.StatusActive, "Trip created", gomock.Any()).After(insert).
					DoAndReturn(func(_ context.Context, id uuid.UUID, _ *string, _, _ string, at time.Time) error {
						if id != inserted.ID || !at.Equal(inserted.StartedAt) {
							t.Fatalf("history does not match inserted trip: %s at %v", id, at)
						}
						return tc.historyErr
					})
			}
			got, err := create.New(repo, tx).Create(ctx, input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Create: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				inserted = trip.Trip{}
			}
			if got != inserted {
				t.Fatalf("Create: got %+v, want %+v", got, inserted)
			}
		})
	}
}

func TestCreateWithKeyReturnsSavedSnapshot(t *testing.T) {
	input := validInput()
	key := uuid.New()
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(body)
	snapshot := trip.Trip{ID: uuid.New(), DriverID: input.DriverID, UserID: input.UserID, Status: trip.StatusActive, Price: input.Price}
	for _, changed := range []bool{false, true} {
		name := "replay"
		if changed {
			name = "different payload"
		}
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo, tx := NewMocktripRepository(ctrl), NewMocktxManager(ctrl)
			ctx := t.Context()
			txCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			tx.EXPECT().Do(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(txCtx)
			})
			gomock.InOrder(
				repo.EXPECT().LockIdempotencyKey(txCtx, key).Return(nil),
				repo.EXPECT().GetIdempotency(txCtx, key).Return(model.IdempotencyRecord{
					RequestHash: hash[:], Trip: snapshot, ExpiresAt: time.Now().Add(time.Hour),
				}, nil),
			)
			request := input
			if changed {
				request.Price++
			}
			got, replayed, err := create.New(repo, tx).CreateWithKey(ctx, request, &key)
			if changed {
				if !errors.Is(err, trip.ErrIdempotencyConflict) || replayed || got != (trip.Trip{}) {
					t.Fatalf("expected idempotency conflict, got %+v, replayed %v, error %v", got, replayed, err)
				}
				return
			}
			if err != nil || !replayed || got != snapshot {
				t.Fatalf("expected original snapshot, got %+v, replayed %v, error %v", got, replayed, err)
			}
		})
	}
}

func TestCreateValidatesBeforeOpeningTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := uuid.New()
	input := validInput()
	input.Price = -1
	got, replayed, err := create.New(NewMocktripRepository(ctrl), NewMocktxManager(ctrl)).CreateWithKey(t.Context(), input, &key)
	var validationErr *trip.ValidationError
	if !errors.As(err, &validationErr) || validationErr.Field != "price" || replayed || got != (trip.Trip{}) {
		t.Fatalf("expected price validation error, got %+v, replayed %v, error %v", got, replayed, err)
	}
}

func validInput() model.CreateTripInput {
	return model.CreateTripInput{
		UserID: uuid.New(), DriverID: uuid.New(), Price: 1450,
		StartPoint: trip.Point{Latitude: 59.9, Longitude: 30.3},
		EndPoint:   trip.Point{Latitude: 60, Longitude: 31},
	}
}
