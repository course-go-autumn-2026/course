package finish_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/finish"
)

func TestFinish(t *testing.T) {
	storageErr := errors.New("storage failure")
	commitErr := errors.New("commit failure")
	cases := []struct {
		name                                   string
		status                                 string
		getErr, saveErr, historyErr, commitErr error
		wantErr                                error
	}{
		{name: "success", status: trip.StatusActive},
		{name: "already completed", status: trip.StatusCompleted, wantErr: trip.ErrCompleted},
		{name: "not found", getErr: trip.ErrNotFound, wantErr: trip.ErrNotFound},
		{name: "read failure", getErr: storageErr, wantErr: storageErr},
		{name: "save failure", status: trip.StatusActive, saveErr: storageErr, wantErr: storageErr},
		{name: "history failure", status: trip.StatusActive, historyErr: storageErr, wantErr: storageErr},
		{name: "commit failure", status: trip.StatusActive, commitErr: commitErr, wantErr: commitErr},
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
			id := uuid.New()
			initial := trip.Trip{ID: id, Status: tc.status}
			load := repo.EXPECT().GetForUpdate(txCtx, id).Return(initial, tc.getErr)
			var saved trip.Trip
			if tc.getErr == nil && tc.status == trip.StatusActive {
				save := repo.EXPECT().SaveCompletion(txCtx, gomock.Any()).After(load).
					DoAndReturn(func(_ context.Context, value trip.Trip) error {
						saved = value
						if saved.ID != id || saved.Status != trip.StatusCompleted || saved.FinishedAt == nil || saved.FinishedAt.IsZero() {
							t.Fatalf("unexpected saved completion: %+v", saved)
						}
						return tc.saveErr
					})
				if tc.saveErr == nil {
					repo.EXPECT().InsertHistory(txCtx, id, &initial.Status, trip.StatusCompleted, "Trip finished", gomock.Any()).After(save).
						DoAndReturn(func(_ context.Context, _ uuid.UUID, _ *string, _, _ string, at time.Time) error {
							if !at.Equal(*saved.FinishedAt) {
								t.Fatalf("history time %v differs from completion time %v", at, saved.FinishedAt)
							}
							return tc.historyErr
						})
				}
			}
			if tc.wantErr == nil {
				events.EXPECT().Completed(ctx, gomock.Any()).Do(func(_ context.Context, id uuid.UUID) {
					if !committed || id != saved.ID {
						t.Fatalf("event before commit or with wrong trip ID: %s", id)
					}
				})
			}
			got, err := finish.New(repo, tx, events).Finish(ctx, id)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Finish: got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				saved = trip.Trip{}
			}
			if got != saved {
				t.Fatalf("Finish: got %+v, want %+v", got, saved)
			}
		})
	}
}
