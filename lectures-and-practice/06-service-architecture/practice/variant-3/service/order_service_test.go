package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/course-go-autumn-2026/practice6-variant3-order/domain"
	"github.com/course-go-autumn-2026/practice6-variant3-order/repository"
	"github.com/course-go-autumn-2026/practice6-variant3-order/service"
)

type stockFunc func(context.Context) (bool, error)

func (f stockFunc) Check(ctx context.Context) (bool, error) { return f(ctx) }

type fraudFunc func(context.Context) (bool, error)

func (f fraudFunc) Check(ctx context.Context) (bool, error) { return f(ctx) }

type recordingOrders struct {
	service.OrderRepository
	events *[]string
}

func (r recordingOrders) Create(ctx context.Context, order *domain.Order) (int64, error) {
	*r.events = append(*r.events, "create")
	return r.OrderRepository.Create(ctx, order)
}

func (r recordingOrders) Update(ctx context.Context, order *domain.Order) error {
	*r.events = append(*r.events, "update")
	return r.OrderRepository.Update(ctx, order)
}

func TestCreateOrder_ExternalChecks(t *testing.T) {
	for _, tc := range []struct {
		name      string
		available bool
		stockErr  error
		blocked   bool
		fraudErr  error
		status    domain.Status
		wantErr   error
	}{
		{"confirmed", true, nil, false, nil, domain.StatusConfirmed, nil},
		{"stock_refused", false, nil, false, nil, domain.StatusPendingStock, nil},
		{"stock_unavailable", true, context.DeadlineExceeded, false, nil, domain.StatusPendingStock, nil},
		{"fraud_unavailable", true, nil, true, context.DeadlineExceeded, domain.StatusConfirmed, nil},
		{"both_unavailable", false, context.DeadlineExceeded, false, context.DeadlineExceeded, domain.StatusPendingStock, nil},
		{"fraud_blocked", true, nil, true, nil, "", domain.ErrPaymentFraud},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			var events []string
			repo := repository.NewMemoryOrderRepository()
			orders := recordingOrders{OrderRepository: repo, events: &events}
			stock := stockFunc(func(context.Context) (bool, error) {
				events = append(events, "stock")
				stored, err := repo.Get(ctx, 1)
				if err != nil || stored.Status != domain.StatusCreated {
					t.Fatalf("order must be saved before stock check: %+v, %v", stored, err)
				}
				return tc.available, tc.stockErr
			})
			fraud := fraudFunc(func(context.Context) (bool, error) {
				events = append(events, "fraud")
				return tc.blocked, tc.fraudErr
			})
			svc := service.NewOrderService(orders, repository.NewMemoryCustomerRepository(), stock, fraud)
			order, err := svc.Create(ctx, service.CreateOrderInput{
				CustomerID: 1,
				Items:      []domain.Item{{SKU: "lamp", Quantity: 2, Price: 500}, {SKU: "book", Quantity: 1, Price: 250}},
			})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if tc.wantErr != nil {
				if !reflect.DeepEqual(events, []string{"fraud"}) {
					t.Fatalf("unexpected calls after blocked payment: %v", events)
				}
				if _, err := repo.Get(ctx, 1); !errors.Is(err, domain.ErrNotFound) {
					t.Fatalf("blocked order must not be stored: %v", err)
				}
				return
			}
			if !reflect.DeepEqual(events, []string{"fraud", "create", "stock", "update"}) {
				t.Fatalf("unexpected call order: %v", events)
			}
			wantQty := 0
			if tc.status == domain.StatusConfirmed {
				wantQty = 3
			}
			if order.Status != tc.status || order.ConfirmedQty != wantQty || order.Total != 1250 {
				t.Fatalf("unexpected order: %+v", order)
			}
			stored, err := svc.Get(ctx, order.ID)
			if err != nil || !reflect.DeepEqual(stored, order) {
				t.Fatalf("saved order differs: %+v, %v", stored, err)
			}
		})
	}
}

func TestCreateOrder_InvalidInputSkipsExternalChecks(t *testing.T) {
	check := stockFunc(func(context.Context) (bool, error) {
		t.Fatal("external checks must not run for invalid input")
		return false, nil
	})
	svc := service.NewOrderService(repository.NewMemoryOrderRepository(), repository.NewMemoryCustomerRepository(), check, check)
	for _, input := range []service.CreateOrderInput{
		{CustomerID: 1},
		{CustomerID: 1, Items: []domain.Item{{Quantity: 0, Price: 1}}},
		{CustomerID: 1, Items: []domain.Item{{Quantity: 1, Price: 100_001}}},
	} {
		if _, err := svc.Create(context.Background(), input); err == nil {
			t.Fatal("want validation error")
		}
	}
}

func TestUpdateItems(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   domain.Status
		quantity int
		price    float64
		want     domain.Status
		wantErr  error
	}{
		{"created", domain.StatusCreated, 3, 500, domain.StatusCreated, nil},
		{"pending", domain.StatusPendingStock, 1, 500, domain.StatusPendingStock, nil},
		{"confirmed_less", domain.StatusConfirmed, 1, 500, domain.StatusConfirmed, nil},
		{"confirmed_equal", domain.StatusConfirmed, 2, 500, domain.StatusConfirmed, nil},
		{"confirmed_more", domain.StatusConfirmed, 3, 500, domain.StatusPendingStock, nil},
		{"cancelled", domain.StatusCancelled, 3, 500, domain.StatusCancelled, domain.ErrOrderCancelled},
		{"no_creation_limit", domain.StatusConfirmed, 1, 2_000_000, domain.StatusConfirmed, nil},
		{"no_creation_quantity_check", domain.StatusCreated, -1, 500, domain.StatusCreated, nil},
		{"no_creation_price_check", domain.StatusCreated, 1, -500, domain.StatusCreated, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			repo := repository.NewMemoryOrderRepository()
			original := &domain.Order{
				CustomerID: 1, Items: []domain.Item{{SKU: "lamp", Quantity: 2, Price: 500}},
				Total: 1000, ConfirmedQty: 2, Status: tc.status,
			}
			id, err := repo.Create(ctx, original)
			if err != nil {
				t.Fatal(err)
			}
			check := stockFunc(func(context.Context) (bool, error) {
				t.Fatal("PATCH must not call external checks")
				return false, nil
			})
			svc := service.NewOrderService(repo, repository.NewMemoryCustomerRepository(), check, check)
			items := []domain.Item{{SKU: "other", Quantity: tc.quantity, Price: tc.price}}
			_, err = svc.UpdateItems(ctx, id, items)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			stored, err := svc.Get(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if stored.Status != tc.want || stored.ConfirmedQty != 2 {
				t.Fatalf("unexpected state: %+v", stored)
			}
			if tc.wantErr != nil {
				if !reflect.DeepEqual(stored.Items, original.Items) || stored.Total != original.Total {
					t.Fatalf("cancelled order changed: %+v", stored)
				}
			} else if !reflect.DeepEqual(stored.Items, items) || stored.Total != tc.price*float64(tc.quantity) {
				t.Fatalf("items or total were not saved: %+v", stored)
			}
		})
	}
}
