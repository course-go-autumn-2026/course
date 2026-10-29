package library

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type LoanPlan interface {
	Name() string
	Price() int
}

type StandardLoan struct{}
type ExtendedLoan struct{}

type WithScanning struct{ Base LoanPlan }
type WithProtectiveCover struct{ Base LoanPlan }

type LoanPrice func(base int) int

var ErrUnknownLoanPlan = errors.New("unknown product")
var ErrUnknownTariff = errors.New("unknown tariff")
var ErrNotRefactored = errors.New("not refactored")
