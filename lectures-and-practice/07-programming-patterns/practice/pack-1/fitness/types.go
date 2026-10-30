package fitness

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Membership interface {
	Name() string
	Price() int
}

type DaytimePass struct{}
type FullDayPass struct{}

type WithLocker struct{ Base Membership }
type WithTowel struct{ Base Membership }

type ActivationPrice func(base int) int

var ErrUnknownMembership = errors.New("unknown product")
var ErrUnknownTariff = errors.New("unknown tariff")
var ErrNotRefactored = errors.New("not refactored")
