package service

import (
	"context"

	"github.com/course-go-autumn-2026/practice6-variant3-order/domain"
)

type OrderService struct {
	orders    OrderRepository
	customers CustomerRepository
	stock     StockChecker
	fraud     PaymentFraudChecker
}

func NewOrderService(orders OrderRepository, customers CustomerRepository, stock StockChecker, fraud PaymentFraudChecker) *OrderService {
	return &OrderService{orders: orders, customers: customers, stock: stock, fraud: fraud}
}

type CreateOrderInput struct {
	CustomerID int64
	Items      []domain.Item
}

func (s *OrderService) Create(ctx context.Context, in CreateOrderInput) (*domain.Order, error) {
	if err := domain.ValidateItems(in.Items); err != nil {
		return nil, err
	}
	customer, err := s.customers.Get(ctx, in.CustomerID)
	if err != nil {
		return nil, err
	}
	order, err := domain.NewOrder(*customer, in.Items)
	if err != nil {
		return nil, err
	}
	blocked, fraudErr := s.fraud.Check(ctx)
	if fraudErr == nil && blocked {
		return nil, domain.ErrPaymentFraud
	}
	id, err := s.orders.Create(ctx, order)
	if err != nil {
		return nil, err
	}
	order.ID = id
	available, stockErr := s.stock.Check(ctx)
	if stockErr != nil || !available {
		err = order.MarkPendingStock()
	} else {
		err = order.ConfirmStock()
	}
	if err != nil {
		return nil, err
	}
	if err := s.orders.Update(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) Get(ctx context.Context, id int64) (*domain.Order, error) {
	return s.orders.Get(ctx, id)
}

func (s *OrderService) UpdateItems(ctx context.Context, id int64, items []domain.Item) (*domain.Order, error) {
	if err := domain.ValidateItemsPresent(items); err != nil {
		return nil, err
	}
	order, err := s.orders.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := order.ChangeItems(items); err != nil {
		return nil, err
	}
	if err := s.orders.Update(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}
