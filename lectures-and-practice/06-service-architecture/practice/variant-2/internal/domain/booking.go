package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidInterval  = errors.New("end_at must be after start_at")
	ErrBookingNotFound  = errors.New("booking not found")
	ErrRoomBooked       = errors.New("room is already booked for this time")
	ErrCancelled        = errors.New("cannot reschedule cancelled booking")
	ErrEmployeeNotFound = errors.New("employee not found")
)

type Status string

const (
	Pending      Status = "pending_confirmation"
	Confirmed    Status = "confirmed"
	Cancelled    Status = "cancelled"
	Rescheduling Status = "rescheduling"
)

type Role string

const (
	Regular Role = "regular"
	Admin   Role = "admin"
)

type Employee struct {
	ID   int64
	Role Role
}

func (e Employee) CanOverrideBooking() bool { return e.Role == Admin }

type Booking struct {
	id, roomID, employeeID    int64
	startAt, endAt, createdAt time.Time
	status                    Status
}

func NewBooking(roomID, employeeID int64, startAt, endAt time.Time, available bool, now time.Time) (Booking, error) {
	if !endAt.After(startAt) {
		return Booking{}, ErrInvalidInterval
	}
	status := Pending
	if available {
		status = Confirmed
	}
	return Booking{
			roomID:     roomID,
			employeeID: employeeID,
			startAt:    startAt,
			endAt:      endAt,
			status:     status,
			createdAt:  now},
		nil
}
func (b Booking) AssignID(id int64) Booking {
	b.id = id
	return b
}
func (b Booking) ID() int64 {
	return b.id
}
func (b Booking) RoomID() int64 {
	return b.roomID
}
func (b Booking) Status() Status {
	return b.status
}
func (b Booking) Overlaps(roomID int64, start, end time.Time) bool {
	return b.roomID == roomID && b.status != Cancelled && start.Before(b.endAt) && b.startAt.Before(end)
}
func (b *Booking) Cancel() { b.status = Cancelled }
func (b *Booking) Reschedule(start, end, now time.Time, confirmWindow time.Duration) error {
	if !end.After(start) {
		return ErrInvalidInterval
	}
	if b.status == Cancelled {
		return ErrCancelled
	}
	b.startAt, b.endAt = start, end
	if b.status == Confirmed && start.Sub(now) < confirmWindow {
		b.status = Rescheduling
	}
	return nil
}

// Snapshot is a response DTO; it retains the original JSON field names.
type Snapshot struct {
	ID         int64
	RoomID     int64
	EmployeeID int64
	StartAt    time.Time
	EndAt      time.Time
	Status     Status
	CreatedAt  time.Time
}

func (b Booking) Snapshot() Snapshot {
	return Snapshot{b.id, b.roomID, b.employeeID, b.startAt, b.endAt, b.status, b.createdAt}
}
