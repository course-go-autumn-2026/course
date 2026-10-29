package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nestera-sh/course_06/domain"
	"github.com/nestera-sh/course_06/repository"
	"github.com/nestera-sh/course_06/service"
)

type fakeModerator struct {
	approved bool
	err      error
	calls    int
}

func (f *fakeModerator) Check(ctx context.Context, id int64) (bool, error) {
	f.calls++
	return f.approved, f.err
}

type fakeAntifraud struct {
	blocked bool
	err     error
}

func (f *fakeAntifraud) Blocked(
	ctx context.Context,
	sellerID int64,
	price float64,
) (bool, error) {
	return f.blocked, f.err
}

func TestCreatePriceAboveLimit(t *testing.T) {
	svc := service.NewListingService(
		repository.NewMemoryListingRepo(),
		&fakeModerator{},
		&fakeAntifraud{},
	)

	_, err := svc.Create(context.Background(), service.CreateInput{
		Title:    "car",
		Price:    11_000_000,
		PhotoURL: "http://x/car.jpg",
		SellerID: 1,
	})

	if !errors.Is(err, domain.ErrPriceSuspicious) {
		t.Fatalf("got error %v, want ErrPriceSuspicious", err)
	}
}

func TestTrustedSellerWithAntifraudError(t *testing.T) {
	moderator := &fakeModerator{}
	repo := repository.NewMemoryListingRepo()
	svc := service.NewListingService(
		repo,
		moderator,
		&fakeAntifraud{err: errors.New("connection refused")},
	)

	listing, err := svc.Create(context.Background(), service.CreateInput{
		Title:    "car",
		Price:    20_000_000,
		PhotoURL: "http://x/car.jpg",
		SellerID: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	if listing.Status != domain.StatusApproved {
		t.Errorf("got status %q, want approved", listing.Status)
	}
	if moderator.calls != 0 {
		t.Errorf("moderation called %d times, want 0", moderator.calls)
	}

	saved, err := repo.Get(context.Background(), listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Status != domain.StatusApproved {
		t.Errorf("saved status = %q, want approved", saved.Status)
	}
}

func TestCreateModerationResults(t *testing.T) {
	tests := []struct {
		name     string
		approved bool
		err      error
		want     domain.Status
	}{
		{"approved", true, nil, domain.StatusApproved},
		{"rejected", false, nil, domain.StatusRejected},
		{"unavailable", false, errors.New("offline"), domain.StatusPending},
		{"timeout", false, context.DeadlineExceeded, domain.StatusReview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewListingService(
				repository.NewMemoryListingRepo(),
				&fakeModerator{approved: tt.approved, err: tt.err},
				&fakeAntifraud{},
			)

			listing, err := svc.Create(context.Background(), service.CreateInput{
				Title:    "lamp",
				Price:    500,
				PhotoURL: "http://x/lamp.jpg",
				SellerID: 1,
			})
			if err != nil {
				t.Fatal(err)
			}

			if listing.Status != tt.want {
				t.Errorf("got status %q, want %q", listing.Status, tt.want)
			}
		})
	}
}

func TestApprovedListingPriceAboveLimit(t *testing.T) {
	svc := service.NewListingService(
		repository.NewMemoryListingRepo(),
		&fakeModerator{approved: true},
		&fakeAntifraud{},
	)
	ctx := context.Background()

	listing, err := svc.Create(ctx, service.CreateInput{
		Title:    "lamp",
		Price:    500,
		PhotoURL: "http://x/lamp.jpg",
		SellerID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if listing.Status != domain.StatusApproved {
		t.Fatalf("initial status = %q, want approved", listing.Status)
	}

	_, err = svc.UpdatePrice(ctx, listing.ID, 11_000_000)
	if err != nil {
		t.Fatal(err)
	}

	saved, err := svc.Get(ctx, listing.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Price != 11_000_000 {
		t.Errorf("got price %g, want 11000000", saved.Price)
	}
	if saved.Status != domain.StatusPending {
		t.Errorf("got status %q, want pending", saved.Status)
	}
}
