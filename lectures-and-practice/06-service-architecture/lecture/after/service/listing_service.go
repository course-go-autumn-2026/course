package service

import (
	"context"
	"errors"

	"github.com/course-go-autumn-2026/lecture6-architecture/after/domain"
)

type ListingService struct {
	repo      ListingRepository
	moderator Moderator
}

func NewListingService(repo ListingRepository, moderator Moderator) *ListingService {
	return &ListingService{repo: repo, moderator: moderator}
}

type PublishInput struct {
	Title    string
	Price    float64
	PhotoURL string
	SellerID int64
}

// Publish - бизнес-сценарий целиком, но без единой строчки SQL
// или net/http. Сравните с before/handler.go: там это была одна
// функция на 80+ строк, здесь - явная последовательность шагов.
func (s *ListingService) Publish(ctx context.Context, in PublishInput) (*domain.Listing, error) {
	listing, err := domain.NewListing(in.Title, in.Price, in.PhotoURL, in.SellerID)
	if err != nil {
		// Ошибка пришла из домена - сервис её не прячет и не оборачивает
		// без причины, просто прокидывает выше. Решать, в какой HTTP-статус
		// её превратить, будет handler.
		return nil, err
	}

	id, err := s.repo.Create(ctx, listing)
	if err != nil {
		return nil, err
	}
	listing.ID = id

	approved, err := s.moderator.Check(ctx, id)
	if err != nil {
		// Модерация недоступна. В before/handler.go это решалось
		// неявно (просто отвечали 201 pending). Здесь это явное
		// решение сервисного слоя: объявление остаётся pending,
		// модерация подхватит его позже (например, по ретраю).
		return listing, nil
	}

	if approved {
		if err := listing.Approve(); err != nil {
			return nil, err
		}
	} else {
		if err := listing.Reject(); err != nil {
			return nil, err
		}
	}

	if err := s.repo.UpdateStatus(ctx, id, listing.Status); err != nil {
		return nil, err
	}
	return listing, nil
}

func (s *ListingService) Get(ctx context.Context, id int64) (*domain.Listing, error) {
	listing, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return listing, nil
}
