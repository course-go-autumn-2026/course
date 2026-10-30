package shipping

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Dispatch struct {
	ID              int64
	Packages, Total int
}

const MaxPackages = 5

var (
	ErrEmpty         = errors.New("no packages")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max packages must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(Dispatch) error

type Event struct {
	DispatchID int64
	Status     string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

func (n *Notifier) Dispatched(o Dispatch) error {
	return n.Publish(Event{DispatchID: o.ID, Status: "dispatched"})
}

type Config struct{ MaxPackages, PickupFee int }
type Option func(*Config) error
type DeliveryService struct{ config Config }

func (p *DeliveryService) Config() Config { return p.config }
func (p *DeliveryService) CanHandle(count int) bool {
	return count > 0 && count <= p.config.MaxPackages
}
func (p *DeliveryService) TotalWithFee(base int) int { return base + p.config.PickupFee }
