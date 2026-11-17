package get

import (
	"context"

	"github.com/google/uuid"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
)

type Usecase struct {
	repository tripRepository
}

func New(repository tripRepository) *Usecase {
	return &Usecase{repository: repository}
}

func (usecase *Usecase) Get(ctx context.Context, id uuid.UUID) (trip.Trip, error) {
	return usecase.repository.GetByID(ctx, id)
}
