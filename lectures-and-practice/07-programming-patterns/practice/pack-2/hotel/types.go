package hotel

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Reservation struct {
	ID            int64
	Nights, Total int
}

const MaxNights = 5

var (
	ErrEmpty         = errors.New("no nights")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max nights must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(Reservation) error

type Event struct {
	ReservationID int64
	Status        string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

func (n *Notifier) Confirmed(o Reservation) error {
	return n.Publish(Event{ReservationID: o.ID, Status: "confirmed"})
}

type Config struct{ MaxNights, ServiceFee int }
type Option func(*Config) error
type Hotel struct{ config Config }

func (p *Hotel) Config() Config { return p.config }
func (p *Hotel) CanHandle(count int) bool {
	return count > 0 && count <= p.config.MaxNights
}
func (p *Hotel) TotalWithFee(base int) int { return base + p.config.ServiceFee }
