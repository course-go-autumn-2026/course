package repository

import (
	"context"
	"sync"

	"github.com/nestera-sh/course_06/domain"
)

type MemoryListingRepo struct {
	mu   sync.RWMutex
	data map[int64]*domain.Listing
	next int64
}

func NewMemoryListingRepo() *MemoryListingRepo {
	return &MemoryListingRepo{
		data: make(map[int64]*domain.Listing),
	}
}

func (r *MemoryListingRepo) Create(ctx context.Context, l *domain.Listing) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.next++
	id := r.next
	l.ID = id

	copyL := *l
	r.data[id] = &copyL

	return id, nil
}

func (r *MemoryListingRepo) Get(ctx context.Context, id int64) (*domain.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	l, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}

	copyL := *l
	return &copyL, nil
}

func (r *MemoryListingRepo) Update(ctx context.Context, l *domain.Listing) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[l.ID]; !ok {
		return domain.ErrNotFound
	}

	copyL := *l
	r.data[l.ID] = &copyL

	return nil
}
