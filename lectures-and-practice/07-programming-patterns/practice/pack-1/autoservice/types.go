package autoservice

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Service interface {
	Name() string
	Price() int
}

type Inspection struct{}
type WheelAlignment struct{}

type WithCarWash struct{ Base Service }
type WithPrioritySlot struct{ Base Service }

type ServicePrice func(base int) int

var ErrUnknownService = errors.New("unknown product")
var ErrUnknownTariff = errors.New("unknown tariff")
var ErrNotRefactored = errors.New("not refactored")
