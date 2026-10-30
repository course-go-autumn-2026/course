package service

import (
	"context"
	"time"

	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/domain"
)

type BookingRepository interface {
	Create(ctx context.Context, b *domain.Booking) (int64, error)
	Get(ctx context.Context, id int64) (*domain.Booking, error)
	Update(ctx context.Context, id int64, b *domain.Booking) (*domain.Booking, error)
	GetBookings(ctx context.Context, roomID int64, startAt time.Time, endAt time.Time) ([]*domain.Booking, error)
}

type EmployeeRepository interface {
	Get(ctx context.Context, id int64) (*domain.Employee, error)
}

type RoomAvailability interface {
	Check(ctx context.Context, id int64, startAt time.Time, endAt time.Time) (bool, error)
}

type Notifier interface {
	BookingCreated(ctx context.Context, id int64) error
}
