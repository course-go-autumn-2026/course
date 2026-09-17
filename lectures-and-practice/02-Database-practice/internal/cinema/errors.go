package cinema

import "errors"

var (
	ErrNoSeats        = errors.New("no available seats")
	ErrAlreadyBooked  = errors.New("user has already booked this screening")
	ErrNotImplemented = errors.New("complete the TODO in BookingService.Book")
)
