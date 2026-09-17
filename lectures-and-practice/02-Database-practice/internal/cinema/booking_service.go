package cinema

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type transactionRunner interface {
	WithinTransaction(context.Context, func(context.Context, pgx.Tx) error) error
}

type screeningSeatTaker interface {
	TakeSeat(context.Context, DBTX, int64) error
}

type bookingCreator interface {
	Create(context.Context, DBTX, int64, int64) (Booking, error)
}

type BookingService struct {
	txManager           transactionRunner
	screeningRepository screeningSeatTaker
	bookingRepository   bookingCreator
}

func NewBookingService(
	txManager transactionRunner,
	screeningRepository screeningSeatTaker,
	bookingRepository bookingCreator,
) *BookingService {
	return &BookingService{
		txManager:           txManager,
		screeningRepository: screeningRepository,
		bookingRepository:   bookingRepository,
	}
}

func (s *BookingService) Book(
	ctx context.Context,
	screeningID int64,
	userID int64,
) (Booking, error) {
	// TODO(задание 4): выполните TakeSeat и Create в одной транзакции.
	return Booking{}, ErrNotImplemented
}
