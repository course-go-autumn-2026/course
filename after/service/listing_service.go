package service

import (
	"context"
	"errors"

	"github.com/nestera-sh/course_06/domain"
)

type ListingService struct {
	repo      ListingRepository
	moderator Moderator
	antifraud Antifraud
}

func NewListingService(r ListingRepository, m Moderator, a Antifraud) *ListingService {
	return &ListingService{
		repo:      r,
		moderator: m,
		antifraud: a,
	}
}

type CreateInput struct {
	Title    string
	Price    float64
	PhotoURL string
	SellerID int64
}

func (s *ListingService) Create(ctx context.Context, in CreateInput) (*domain.Listing, error) {
	listing, err := domain.NewListing(in.Title, in.Price, in.PhotoURL, in.SellerID)
	if err != nil {
		return nil, err
	}

	// Антифрод (fail-open: ошибку игнорируем)
	blocked, _ := s.antifraud.Blocked(ctx, in.SellerID, in.Price)

	// Сохраняем
	id, err := s.repo.Create(ctx, listing)
	if err != nil {
		return nil, err
	}
	listing.ID = id

	// Если заблокирован антифродом
	if blocked {
		listing.Reject()
		_ = s.repo.Update(ctx, listing)
		return listing, nil
	}

	// Доверенный продавец минует модерацию (пока хардкод)
	if in.SellerID == 2 {
		listing.Approve()
		_ = s.repo.Update(ctx, listing)
		return listing, nil
	}

	// Модерация (fail-safe)
	approved, modErr := s.moderator.Check(ctx, id)
	switch {
	case errors.Is(modErr, context.DeadlineExceeded):
		listing.MarkReview() // таймаут -> review
	case modErr != nil:
		listing.MarkPending() // сетевая ошибка -> pending
	case approved:
		listing.Approve()
	default:
		listing.Reject()
	}

	_ = s.repo.Update(ctx, listing)
	return listing, nil
}

func (s *ListingService) Get(ctx context.Context, id int64) (*domain.Listing, error) {
	return s.repo.Get(ctx, id)
}

func (s *ListingService) UpdatePrice(ctx context.Context, id int64, newPrice float64) (*domain.Listing, error) {
	listing, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := listing.ChangePrice(newPrice); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, listing); err != nil {
		return nil, err
	}

	return listing, nil
}
