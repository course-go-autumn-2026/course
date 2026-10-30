package service

import (
	"context"
	"errors"
	"time"

	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/domain"
)

// Перенос confirmed-брони ближе этого срока до начала требует
// повторного подтверждения.
const rescheduleConfirmWindow = 2 * time.Hour

type BookingService struct {
	bookingRepo  BookingRepository
	employeeRepo EmployeeRepository
	availability RoomAvailability
	notifier     Notifier
}

func NewBookingService(
	bookings BookingRepository,
	employees EmployeeRepository,
	availability RoomAvailability,
	notifier Notifier,
) *BookingService {
	return &BookingService{
		bookingRepo:  bookings,
		employeeRepo: employees,
		availability: availability,
		notifier:     notifier,
	}
}

type CreateBookingInput struct {
	RoomID     int64
	EmployeeID int64
	StartAt    time.Time
	EndAt      time.Time
}

func (s *BookingService) Create(ctx context.Context, in CreateBookingInput) (*domain.Booking, error) {
	if !in.EndAt.After(in.StartAt) {
		return nil, domain.ErrInvalidInterval
	}

	canOverride, err := s.canOverride(ctx, in.EmployeeID)
	if err != nil {
		return nil, err
	}

	existing, err := s.bookingRepo.GetBookings(ctx, in.RoomID, in.StartAt, in.EndAt)
	if err != nil {
		return nil, err
	}

	var conflicts []*domain.Booking
	for _, b := range existing {
		if b.Overlaps(in.RoomID, in.StartAt, in.EndAt) {
			conflicts = append(conflicts, b)
		}
	}

	if len(conflicts) > 0 {
		if !canOverride {
			return nil, domain.ErrRoomBooked
		}
		// Админ побеждает: отменяются все пересекающиеся брони.
		for _, c := range conflicts {
			c.Cancel()
			if _, err := s.bookingRepo.Update(ctx, c.ID(), c); err != nil {
				return nil, err
			}
		}
	}

	available, err := s.availability.Check(ctx, in.RoomID, in.StartAt, in.EndAt)
	if err != nil {
		available = false
	}

	booking, err := domain.NewBooking(in.RoomID, in.EmployeeID, in.StartAt, in.EndAt, available, time.Now())
	if err != nil {
		return nil, err
	}

	id, err := s.bookingRepo.Create(ctx, &booking)
	if err != nil {
		return nil, err
	}
	booking = booking.AssignID(id) // AssignID возвращает копию, результат надо сохранить

	// Fail-open: рассылка не блокирует создание и не влияет на статус.
	_ = s.notifier.BookingCreated(ctx, id)

	return &booking, nil
}

func (s *BookingService) Get(ctx context.Context, id int64) (*domain.Booking, error) {
	return s.bookingRepo.Get(ctx, id)
}

func (s *BookingService) Reschedule(ctx context.Context, id int64, startAt, endAt time.Time) (*domain.Booking, error) {
	// Валидация до обращения к хранилищу: кривой интервал даёт 400
	// даже для несуществующей брони.
	if !endAt.After(startAt) {
		return nil, domain.ErrInvalidInterval
	}

	booking, err := s.bookingRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := booking.Reschedule(startAt, endAt, time.Now(), rescheduleConfirmWindow); err != nil {
		return nil, err
	}
	return s.bookingRepo.Update(ctx, id, booking)
}

// Неизвестный сотрудник считается обычным (regular).
func (s *BookingService) canOverride(ctx context.Context, employeeID int64) (bool, error) {
	emp, err := s.employeeRepo.Get(ctx, employeeID)
	if errors.Is(err, domain.ErrEmployeeNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return emp.CanOverrideBooking(), nil
}
