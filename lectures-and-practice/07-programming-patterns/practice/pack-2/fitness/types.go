package fitness

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Visit struct {
	ID               int64
	Attendees, Total int
}

const MaxAttendees = 5

var (
	ErrEmpty         = errors.New("no attendees")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max attendees must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(Visit) error

type Event struct {
	VisitID int64
	Status  string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

func (n *Notifier) Registered(o Visit) error {
	return n.Publish(Event{VisitID: o.ID, Status: "registered"})
}

type Config struct{ MaxAttendees, RegistrationFee int }
type Option func(*Config) error
type Club struct{ config Config }

func (p *Club) Config() Config { return p.config }
func (p *Club) CanHandle(count int) bool {
	return count > 0 && count <= p.config.MaxAttendees
}
func (p *Club) TotalWithFee(base int) int { return base + p.config.RegistrationFee }
