package cinema

// Готовые контракты. Денежные значения - целые рубли.
import "errors"

type Ticket interface {
	Name() string
	Price() int
}

type StandardTicket struct{}
type PremiumTicket struct{}

type WithPopcorn struct{ Base Ticket }
type WithDrink struct{ Base Ticket }

type TicketPrice func(base int) int

var ErrUnknownTicket = errors.New("unknown product")
var ErrUnknownTariff = errors.New("unknown tariff")
var ErrNotRefactored = errors.New("not refactored")
