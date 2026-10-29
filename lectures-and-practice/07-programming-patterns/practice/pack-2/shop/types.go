package shop

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Purchase struct {
	ID           int64
	Items, Total int
}

const MaxItems = 5

var (
	ErrEmpty         = errors.New("no items")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max items must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(Purchase) error

type Event struct {
	PurchaseID int64
	Status     string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

func (n *Notifier) Shipped(o Purchase) error {
	return n.Publish(Event{PurchaseID: o.ID, Status: "shipped"})
}

type Config struct{ MaxItems, DeliveryFee int }
type Option func(*Config) error
type Shop struct{ config Config }

func (p *Shop) Config() Config { return p.config }
func (p *Shop) CanHandle(count int) bool {
	return count > 0 && count <= p.config.MaxItems
}
func (p *Shop) TotalWithFee(base int) int { return base + p.config.DeliveryFee }
