package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/course-go-autumn-2026/practice6-variant3-order/domain"
	"github.com/course-go-autumn-2026/practice6-variant3-order/repository"
)

func TestMemoryOrderRepository_Snapshots(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryOrderRepository()
	order, err := domain.NewOrder(domain.Customer{ID: 1}, []domain.Item{{SKU: "lamp", Quantity: 2, Price: 500}})
	if err != nil {
		t.Fatal(err)
	}
	id, err := repo.Create(ctx, order)
	if err != nil {
		t.Fatal(err)
	}
	order.Items[0].Price = 999
	stored, err := repo.Get(ctx, id)
	if err != nil || stored.Items[0].Price != 500 {
		t.Fatalf("create retained caller-owned items: %+v, %v", stored, err)
	}
	stored.Items[0].Price = 700
	unchanged, err := repo.Get(ctx, id)
	if err != nil || unchanged.Items[0].Price != 500 {
		t.Fatalf("get exposed stored items: %+v, %v", unchanged, err)
	}
	if err := repo.Update(ctx, stored); err != nil {
		t.Fatal(err)
	}
	stored.Items[0].Price = 888
	updated, err := repo.Get(ctx, id)
	if err != nil || updated.Items[0].Price != 700 {
		t.Fatalf("update retained caller-owned items: %+v, %v", updated, err)
	}
	if _, err := repo.Get(ctx, 999); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if err := repo.Update(ctx, &domain.Order{ID: 999}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestMemoryCustomerRepository(t *testing.T) {
	repo := repository.NewMemoryCustomerRepository()
	for _, tc := range []struct {
		id    int64
		level domain.CustomerLevel
	}{
		{1, domain.CustomerRegular},
		{2, domain.CustomerVIP},
		{999, domain.CustomerNew},
	} {
		customer, err := repo.Get(context.Background(), tc.id)
		if err != nil || customer.ID != tc.id || customer.Level != tc.level {
			t.Fatalf("unexpected customer: %+v, %v", customer, err)
		}
	}
}
