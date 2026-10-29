package service

import (
	"context"
	"github.com/course-go-autumn-2026/lecture6-architecture/after/domain"
)

type ListingRepository interface {
	Create(ctx context.Context, l *domain.Listing) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status domain.Status) error
	Get(ctx context.Context, id int64) (*domain.Listing, error)
}

type Moderator interface {
	Check(ctx context.Context, listingID int64) (approved bool, err error)
}
