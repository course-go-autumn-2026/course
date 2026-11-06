package create_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
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
			events := NewMocktripEvents(ctrl)
			committed := false
			ctx := t.Context()
			txCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			tx.EXPECT().Do(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				if err := fn(txCtx); err != nil {
					return err
				}
				committed = tc.commitErr == nil
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
			if tc.wantErr == nil {
				events.EXPECT().Created(ctx, gomock.Any()).Do(func(_ context.Context, id uuid.UUID) {
					if !committed || id != inserted.ID {
						t.Fatalf("event before commit or with wrong trip ID: %s", id)
					}
				})
			}
			got, err := create.New(repo, tx, events).Create(ctx, input)
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
			got, replayed, err := create.New(repo, tx, NewMocktripEvents(ctrl)).CreateWithKey(ctx, request, &key)
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

func TestCreateWithKeyStoresNewSnapshot(t *testing.T) {
	input := validInput()
	key := uuid.New()
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(body)
	expired := model.IdempotencyRecord{
		RequestHash: make([]byte, sha256.Size),
		Trip:        trip.Trip{ID: uuid.New(), Status: trip.StatusCompleted},
		ExpiresAt:   time.Now().Add(-time.Hour),
	}
	deleteErr := errors.New("delete expired key failed")
	for _, tc := range []struct {
		name      string
		record    model.IdempotencyRecord
		lookupErr error
		deleteErr error
	}{
		{name: "new key", lookupErr: fmt.Errorf("lookup key: %w", trip.ErrNotFound)},
		{name: "expired key with different payload", record: expired},
		{name: "expired key deletion failed", record: expired, deleteErr: deleteErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo, tx := NewMocktripRepository(ctrl), NewMocktxManager(ctrl)
			ctx := t.Context()
			txCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			tx.EXPECT().Do(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, fn func(context.Context) error) error {
				return fn(txCtx)
			})
			lock := repo.EXPECT().LockIdempotencyKey(txCtx, key).Return(nil)
			previous := repo.EXPECT().GetIdempotency(txCtx, key).After(lock).Return(tc.record, tc.lookupErr)
			if tc.lookupErr == nil {
				previous = repo.EXPECT().DeleteIdempotency(txCtx, key).After(previous).Return(tc.deleteErr)
			}
			beforeCreate := time.Now()
			var inserted trip.Trip
			if tc.deleteErr == nil {
				insert := repo.EXPECT().Insert(txCtx, gomock.Any()).After(previous).Do(func(_ context.Context, value trip.Trip) {
					inserted = value
				})
				history := repo.EXPECT().InsertHistory(txCtx, gomock.Any(), (*string)(nil), trip.StatusActive, "Trip created", gomock.Any()).After(insert).Return(nil)
				repo.EXPECT().InsertIdempotency(txCtx, key, hash, gomock.Any(), gomock.Any()).After(history).
					Do(func(_ context.Context, _ uuid.UUID, _ [32]byte, snapshot trip.Trip, expiresAt time.Time) {
						if snapshot != inserted {
							t.Fatalf("snapshot %+v differs from inserted trip %+v", snapshot, inserted)
						}
						savedAt := expiresAt.Add(-24 * time.Hour)
						if savedAt.Before(beforeCreate) || savedAt.After(time.Now()) {
							t.Fatalf("key must expire 24 hours after creation, got %v", expiresAt)
						}
					})
			}
			got, replayed, err := create.New(repo, tx, nil).CreateWithKey(ctx, input, &key)
			if !errors.Is(err, tc.deleteErr) || replayed || got != inserted {
				t.Fatalf("CreateWithKey: got %+v, replayed %t, error %v, want %+v, false, %v", got, replayed, err, inserted, tc.deleteErr)
			}
			if tc.deleteErr == nil && (got.ID == uuid.Nil || got.ID == expired.Trip.ID) {
				t.Fatalf("expected a new trip, got %+v", got)
			}
		})
	}
}

func TestCreateValidatesBeforeOpeningTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := uuid.New()
	input := validInput()
	input.Price = -1
	got, replayed, err := create.New(NewMocktripRepository(ctrl), NewMocktxManager(ctrl), nil).CreateWithKey(t.Context(), input, &key)
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
