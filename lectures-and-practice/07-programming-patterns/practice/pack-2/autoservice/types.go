package autoservice

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type WorkOrder struct {
	ID          int64
	Jobs, Total int
}

const MaxJobs = 5

var (
	ErrEmpty         = errors.New("no jobs")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max jobs must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(WorkOrder) error

type Event struct {
	WorkOrderID int64
	Status      string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

func (n *Notifier) Completed(o WorkOrder) error {
	return n.Publish(Event{WorkOrderID: o.ID, Status: "completed"})
}

type Config struct{ MaxJobs, CalloutFee int }
type Option func(*Config) error
type Workshop struct{ config Config }

func (p *Workshop) Config() Config { return p.config }
func (p *Workshop) CanHandle(count int) bool {
	return count > 0 && count <= p.config.MaxJobs
}
func (p *Workshop) TotalWithFee(base int) int { return base + p.config.CalloutFee }
