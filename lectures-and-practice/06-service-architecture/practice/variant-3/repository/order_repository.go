package repository

import (
	"context"
	"sync"
	"time"

	"github.com/course-go-autumn-2026/practice6-variant3-order/domain"
)

type storedOrder struct {
	ID           int64
	CustomerID   int64
	Items        []domain.Item
	Total        float64
	ConfirmedQty int
	Status       string
	CreatedAt    time.Time
}

type MemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[int64]storedOrder
	nextID int64
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{orders: make(map[int64]storedOrder)}
}

func (r *MemoryOrderRepository) Create(_ context.Context, order *domain.Order) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	stored := storeOrder(order)
	stored.ID = r.nextID
	r.orders[r.nextID] = stored
	return r.nextID, nil
}

func (r *MemoryOrderRepository) Get(_ context.Context, id int64) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stored, ok := r.orders[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &domain.Order{
		ID: stored.ID, CustomerID: stored.CustomerID,
		Items: append([]domain.Item(nil), stored.Items...), Total: stored.Total,
		ConfirmedQty: stored.ConfirmedQty, Status: domain.Status(stored.Status),
		CreatedAt: stored.CreatedAt,
	}, nil
}

func (r *MemoryOrderRepository) Update(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.orders[order.ID]; !ok {
		return domain.ErrNotFound
	}
	r.orders[order.ID] = storeOrder(order)
	return nil
}

func storeOrder(order *domain.Order) storedOrder {
	return storedOrder{
		ID: order.ID, CustomerID: order.CustomerID,
		Items: append([]domain.Item(nil), order.Items...), Total: order.Total,
		ConfirmedQty: order.ConfirmedQty, Status: string(order.Status),
		CreatedAt: order.CreatedAt,
	}
}
