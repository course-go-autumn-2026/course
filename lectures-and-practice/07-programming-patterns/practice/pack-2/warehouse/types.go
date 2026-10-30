package warehouse

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Transfer struct {
	ID            int64
	Places, Total int
}

const MaxPlaces = 5

var (
	ErrEmpty         = errors.New("no places")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max places must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(Transfer) error

type Event struct {
	TransferID int64
	Status     string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

func (n *Notifier) Dispatched(o Transfer) error {
	return n.Publish(Event{TransferID: o.ID, Status: "dispatched"})
}

type Config struct{ MaxPlaces, HandlingFee int }
type Option func(*Config) error
type Warehouse struct{ config Config }

func (p *Warehouse) Config() Config { return p.config }
func (p *Warehouse) CanHandle(count int) bool {
	return count > 0 && count <= p.config.MaxPlaces
}
func (p *Warehouse) TotalWithFee(base int) int { return base + p.config.HandlingFee }
