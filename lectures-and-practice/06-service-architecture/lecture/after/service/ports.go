// Package service - бизнес-сценарии. Знает про domain, но не знает,
// как именно Listing хранится или как устроена модерация - только
// интерфейсы-порты, описанные здесь же.
package service

import (
	"context"

	"github.com/course-go-autumn-2026/lecture6-architecture/after/domain"
)

// ListingRepository - порт для хранилища. Service зависит от этого
// интерфейса, а не от *sql.DB и не от конкретной СУБД. Реализация
// (адаптер) лежит в repository и может быть заменена без изменений
// в service.
type ListingRepository interface {
	Create(ctx context.Context, l *domain.Listing) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status domain.Status) error
	Get(ctx context.Context, id int64) (*domain.Listing, error)
}

// Moderator - порт для внешнего сервиса модерации. Та же идея:
// сегодня это HTTP-вызов, завтра может стать очередью - service
// об этом не узнает.
type Moderator interface {
	Check(ctx context.Context, listingID int64) (approved bool, err error)
}
