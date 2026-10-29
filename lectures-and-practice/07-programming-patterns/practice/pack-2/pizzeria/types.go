package pizzeria

import "errors"

type Order struct {
	ID            int64
	Pizzas, Total int
}

const MaxPizzas = 5

var (
	ErrEmpty         = errors.New("no pizzas")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max pizzas must be positive")
	ErrInvalidFee    = errors.New("delivery fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(Order) error

type Event struct {
	OrderID int64
	Status  string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

// Готовый прикладной вызов: студент реализует Subscribe и Publish.
func (n *Notifier) Ready(o Order) error {
	return n.Publish(Event{OrderID: o.ID, Status: "ready"})
}

// Цены — целые рубли. Настройки возвращаются по значению.
type Config struct{ MaxPizzas, DeliveryFee int }
type Option func(*Config) error
type Pizzeria struct{ config Config }

func (p *Pizzeria) Config() Config { return p.config }
func (p *Pizzeria) CanHandle(pizzas int) bool {
	return pizzas > 0 && pizzas <= p.config.MaxPizzas
}
func (p *Pizzeria) DeliveryTotal(base int) int { return base + p.config.DeliveryFee }
