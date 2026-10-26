package service

import (
	"context"

	"github.com/course-go-autumn-2026/practice6-variant3-order/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) (int64, error)
	Get(ctx context.Context, id int64) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
}

type CustomerRepository interface {
	Get(ctx context.Context, id int64) (*domain.Customer, error)
}

type StockChecker interface {
	Check(ctx context.Context) (available bool, err error)
}

type PaymentFraudChecker interface {
	Check(ctx context.Context) (blocked bool, err error)
}
