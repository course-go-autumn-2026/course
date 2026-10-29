package service

import (
	"context"

	"github.com/nestera-sh/course_06/domain"
)

type ListingRepository interface {
	Create(ctx context.Context, l *domain.Listing) (int64, error)
	Get(ctx context.Context, id int64) (*domain.Listing, error)
	Update(ctx context.Context, l *domain.Listing) error
}

// Moderator — fail-safe порт.
type Moderator interface {
	Check(ctx context.Context, id int64) (approved bool, err error)
}

// Antifraud — fail-open порт.
type Antifraud interface {
	Blocked(ctx context.Context, sellerID int64, price float64) (blocked bool, err error)
}
