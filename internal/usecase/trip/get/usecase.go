package get

import (
	"context"

	"github.com/google/uuid"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
)

type Usecase struct {
	repo tripRepository
}

func New(repo tripRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Get(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	return u.repo.Get(ctx, id)
}
