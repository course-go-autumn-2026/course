package hotel

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Room interface {
	Name() string
	Price() int
}

type StandardRoom struct{}
type SuiteRoom struct{}

type WithBreakfast struct{ Base Room }
type WithParking struct{ Base Room }

type BookingPrice func(base int) int

var ErrUnknownRoom = errors.New("unknown product")
var ErrUnknownTariff = errors.New("unknown tariff")
var ErrNotRefactored = errors.New("not refactored")
