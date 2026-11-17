package domain_test

import (
	"errors"
	"testing"

	"github.com/course-go-autumn-2026/practice6-variant3-order/domain"
)

func TestNewOrder(t *testing.T) {
	for _, tc := range []struct {
		name  string
		level domain.CustomerLevel
		items []domain.Item
		want  error
	}{
		{"empty", domain.CustomerRegular, nil, domain.ErrItemsMissing},
		{"zero_quantity", domain.CustomerRegular, []domain.Item{{Quantity: 0, Price: 1}}, domain.ErrInvalidQuantity},
		{"negative_quantity", domain.CustomerRegular, []domain.Item{{Quantity: -1, Price: 1}}, domain.ErrInvalidQuantity},
		{"zero_price", domain.CustomerRegular, []domain.Item{{Quantity: 1, Price: 0}}, domain.ErrInvalidPrice},
		{"negative_price", domain.CustomerRegular, []domain.Item{{Quantity: 1, Price: -1}}, domain.ErrInvalidPrice},
		{"regular_at_limit", domain.CustomerRegular, []domain.Item{{Quantity: 2, Price: 50_000}}, nil},
		{"regular_over_limit", domain.CustomerRegular, []domain.Item{{Quantity: 2, Price: 50_001}}, domain.ErrOrderLimit},
		{"new_over_limit", domain.CustomerNew, []domain.Item{{Quantity: 1, Price: 100_001}}, domain.ErrOrderLimit},
		{"vip_at_limit", domain.CustomerVIP, []domain.Item{{Quantity: 2, Price: 500_000}}, nil},
		{"vip_over_limit", domain.CustomerVIP, []domain.Item{{Quantity: 1, Price: 1_000_001}}, domain.ErrOrderLimit},
		{"invalid_second_item", domain.CustomerRegular, []domain.Item{{Quantity: 1, Price: 1}, {Quantity: 0, Price: 1}}, domain.ErrInvalidQuantity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			order, err := domain.NewOrder(domain.Customer{ID: 7, Level: tc.level}, tc.items)
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
			if err == nil && (order.CustomerID != 7 || order.Status != domain.StatusCreated || order.CreatedAt.IsZero()) {
				t.Fatalf("unexpected order: %+v", order)
			}
		})
	}
}

func TestOrderStockTransitions(t *testing.T) {
	order, err := domain.NewOrder(domain.Customer{ID: 1}, []domain.Item{
		{SKU: "lamp", Quantity: 2, Price: 500},
		{SKU: "book", Quantity: 1, Price: 250},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := order.ConfirmStock(); err != nil {
		t.Fatal(err)
	}
	if order.Status != domain.StatusConfirmed || order.ConfirmedQty != 3 || order.Total != 1250 {
		t.Fatalf("unexpected confirmed order: %+v", order)
	}
	if err := order.ConfirmStock(); !errors.Is(err, domain.ErrInvalidStatus) {
		t.Fatalf("want invalid transition, got %v", err)
	}
	if err := order.MarkPendingStock(); err != nil {
		t.Fatal(err)
	}
	if err := order.ConfirmStock(); !errors.Is(err, domain.ErrInvalidStatus) {
		t.Fatalf("want invalid transition from pending_stock, got %v", err)
	}
}
