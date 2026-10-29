package courses

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Enrollment struct {
	ID              int64
	Students, Total int
}

const MaxStudents = 5

var (
	ErrEmpty         = errors.New("no students")
	ErrTotal         = errors.New("non-positive total")
	ErrApproval      = errors.New("manager approval required")
	ErrInvalidMax    = errors.New("max students must be positive")
	ErrInvalidFee    = errors.New("fee must not be negative")
	ErrNotRefactored = errors.New("not refactored")
)

type Check func(Enrollment) error

type Event struct {
	EnrollmentID int64
	Status       string
}
type Observer interface{ Update(Event) error }
type ObserverFunc func(Event) error

func (f ObserverFunc) Update(e Event) error { return f(e) }

type Notifier struct{ observers []Observer }

func (n *Notifier) Activated(o Enrollment) error {
	return n.Publish(Event{EnrollmentID: o.ID, Status: "activated"})
}

type Config struct{ MaxStudents, EnrollmentFee int }
type Option func(*Config) error
type School struct{ config Config }

func (p *School) Config() Config { return p.config }
func (p *School) CanHandle(count int) bool {
	return count > 0 && count <= p.config.MaxStudents
}
func (p *School) TotalWithFee(base int) int { return base + p.config.EnrollmentFee }
