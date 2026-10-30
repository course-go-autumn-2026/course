package service_test

import (
	"context"
	"testing"

	"github.com/course-go-autumn-2026/lecture6-architecture/after/domain"
	"github.com/course-go-autumn-2026/lecture6-architecture/after/service"
)

// fakeRepo и fakeModerator - рукописные фейки на портах из ports.go.
// Ни одного SQL-запроса, ни одного HTTP-вызова: тест проверяет только
// бизнес-сценарий. Так выглядит главная практическая польза слоёв -
// before/handler.go таким способом не протестировать вообще,
// там Publish неотделим от *sql.DB и реального HTTP-клиента.
type fakeRepo struct {
	created *domain.Listing
}

func (f *fakeRepo) Create(_ context.Context, l *domain.Listing) (int64, error) {
	f.created = l
	return 42, nil
}

func (f *fakeRepo) UpdateStatus(_ context.Context, _ int64, status domain.Status) error {
	f.created.Status = status
	return nil
}

func (f *fakeRepo) Get(_ context.Context, id int64) (*domain.Listing, error) {
	if f.created == nil || f.created.ID != id {
		return nil, domain.ErrNotFound
	}
	return f.created, nil
}

type fakeModerator struct {
	approve bool
	err     error
}

func (f *fakeModerator) Check(_ context.Context, _ int64) (bool, error) {
	return f.approve, f.err
}

func TestPublish_Approved(t *testing.T) {
	repo := &fakeRepo{}
	svc := service.NewListingService(repo, &fakeModerator{approve: true})

	listing, err := svc.Publish(context.Background(), service.PublishInput{
		Title: "lamp", Price: 500, PhotoURL: "http://x/1.jpg", SellerID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if listing.Status != domain.StatusApproved {
		t.Errorf("want approved, got %s", listing.Status)
	}
}

func TestPublish_InvalidPrice(t *testing.T) {
	svc := service.NewListingService(&fakeRepo{}, &fakeModerator{approve: true})

	_, err := svc.Publish(context.Background(), service.PublishInput{
		Title: "lamp", Price: 0, PhotoURL: "http://x/1.jpg", SellerID: 1,
	})
	if err != domain.ErrInvalidPrice {
		t.Errorf("want ErrInvalidPrice, got %v", err)
	}
}

func TestPublish_ModerationDown_StaysPending(t *testing.T) {
	repo := &fakeRepo{}
	svc := service.NewListingService(repo, &fakeModerator{err: context.DeadlineExceeded})

	listing, err := svc.Publish(context.Background(), service.PublishInput{
		Title: "lamp", Price: 500, PhotoURL: "http://x/1.jpg", SellerID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Явное решение сервисного слоя из listing_service.go: недоступная
	// модерация не валит публикацию, объявление остаётся pending.
	if listing.Status != domain.StatusPending {
		t.Errorf("want pending when moderation is down, got %s", listing.Status)
	}
}
