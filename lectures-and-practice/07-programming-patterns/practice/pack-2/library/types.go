package library

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type LoanRequest struct {
	ID            int64
	Copies, Total int
}

const MaxCopies = 5

var (
	ErrEmpty         = errors.New("no copies")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max copies must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(LoanRequest) error

type Event struct {
	LoanRequestID int64
	Status        string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

func (n *Notifier) Issued(o LoanRequest) error {
	return n.Publish(Event{LoanRequestID: o.ID, Status: "issued"})
}

type Config struct{ MaxCopies, DeliveryFee int }
type Option func(*Config) error
type Library struct{ config Config }

func (p *Library) Config() Config { return p.config }
func (p *Library) CanHandle(count int) bool {
	return count > 0 && count <= p.config.MaxCopies
}
func (p *Library) TotalWithFee(base int) int { return base + p.config.DeliveryFee }
