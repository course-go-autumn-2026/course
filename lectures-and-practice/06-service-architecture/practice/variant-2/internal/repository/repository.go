package repository

import (
	"context"
	"sync"
	"time"

	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/domain"
)

type MemoryBookingRepository struct {
	mu     sync.Mutex
	data   map[int64]domain.Booking // храним по значению для защиты от внешних мутаций
	nextID int64
}

func NewMemoryBookingRepository() *MemoryBookingRepository {
	return &MemoryBookingRepository{
		data: make(map[int64]domain.Booking),
	}
}

func (r *MemoryBookingRepository) Create(ctx context.Context, b *domain.Booking) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	id := r.nextID

	// Используем штатный метод AssignID из domain
	withID := b.AssignID(id)
	r.data[id] = withID

	return id, nil
}

func (r *MemoryBookingRepository) Get(ctx context.Context, id int64) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.data[id]
	if !ok {
		return nil, domain.ErrBookingNotFound
	}

	copied := b
	return &copied, nil
}

func (r *MemoryBookingRepository) Update(ctx context.Context, id int64, b *domain.Booking) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[id]; !ok {
		return nil, domain.ErrBookingNotFound
	}

	withID := b.AssignID(id)
	r.data[id] = withID

	copied := withID
	return &copied, nil
}

func (r *MemoryBookingRepository) GetBookings(ctx context.Context, roomID int64, startAt, endAt time.Time) ([]*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*domain.Booking
	for _, b := range r.data {
		// Используем штатный метод Overlaps из domain
		if b.Overlaps(roomID, startAt, endAt) {
			copied := b
			result = append(result, &copied)
		}
	}
	return result, nil
}

// --- Хранилище сотрудников ---

type storedEmployee struct {
	ID   int64
	Role string
}

type MemoryEmployeeRepository struct {
	mu        sync.RWMutex
	employees map[int64]storedEmployee
}

func NewMemoryEmployeeRepository() *MemoryEmployeeRepository {
	return &MemoryEmployeeRepository{
		// Предзаполняем сотрудников точно как в исходном main.go
		employees: map[int64]storedEmployee{
			1: {ID: 1, Role: string(domain.Regular)},
			2: {ID: 2, Role: string(domain.Admin)},
		},
	}
}

func (r *MemoryEmployeeRepository) Get(ctx context.Context, id int64) (*domain.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.employees[id]
	if !ok {
		return nil, domain.ErrEmployeeNotFound
	}

	return &domain.Employee{
		ID:   e.ID,
		Role: domain.Role(e.Role),
	}, nil
}
