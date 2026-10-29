package cinema

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Booking struct {
	ID           int64
	Seats, Total int
}

const MaxSeats = 5

var (
	ErrEmpty         = errors.New("no seats")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max seats must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(Booking) error

type Event struct {
	BookingID int64
	Status    string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

func (n *Notifier) Confirmed(o Booking) error {
	return n.Publish(Event{BookingID: o.ID, Status: "confirmed"})
}

type Config struct{ MaxSeats, BookingFee int }
type Option func(*Config) error
type Cinema struct{ config Config }

func (p *Cinema) Config() Config { return p.config }
func (p *Cinema) CanHandle(count int) bool {
	return count > 0 && count <= p.config.MaxSeats
}
func (p *Cinema) TotalWithFee(base int) int { return base + p.config.BookingFee }
