package repository

import (
	"context"

	"github.com/course-go-autumn-2026/practice6-variant3-order/domain"
)

type MemoryCustomerRepository struct {
	levels map[int64]domain.CustomerLevel
}

func NewMemoryCustomerRepository() *MemoryCustomerRepository {
	return &MemoryCustomerRepository{
		levels: map[int64]domain.CustomerLevel{1: domain.CustomerRegular, 2: domain.CustomerVIP},
	}
}

func (r *MemoryCustomerRepository) Get(_ context.Context, id int64) (*domain.Customer, error) {
	level, ok := r.levels[id]
	if !ok {
		level = domain.CustomerNew
	}
	return &domain.Customer{ID: id, Level: level}, nil
}
