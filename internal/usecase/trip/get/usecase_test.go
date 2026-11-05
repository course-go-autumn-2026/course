package get_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/get"
)

func TestGet(t *testing.T) {
	id := uuid.New()
	cases := []struct {
		name  string
		value trip.Trip
		err   error
	}{
		{"found", trip.Trip{ID: id, Status: trip.StatusActive}, nil},
		{"missing", trip.Trip{}, trip.ErrNotFound},
		{"repository failure", trip.Trip{}, errors.New("storage failure")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewMocktripRepository(gomock.NewController(t))
			ctx := t.Context()
			repo.EXPECT().Get(ctx, id).Return(tc.value, tc.err)
			got, err := get.New(repo).Get(ctx, id)
			if !errors.Is(err, tc.err) || got != tc.value {
				t.Fatalf("Get: got %+v, %v, want %+v, %v", got, err, tc.value, tc.err)
			}
		})
	}
}
