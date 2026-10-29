package courses

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Course interface {
	Name() string
	Price() int
}

type RecordedCourse struct{}
type LiveCourse struct{}

type WithMentoring struct{ Base Course }
type WithCertificate struct{ Base Course }

type EnrollmentPrice func(base int) int

var ErrUnknownCourse = errors.New("unknown product")
var ErrUnknownTariff = errors.New("unknown tariff")
var ErrNotRefactored = errors.New("not refactored")
