package domain

import "errors"

var (
	ErrInvalid     = errors.New("invalid request")
	ErrOutside     = errors.New("outside delivery area")
	ErrUnavailable = errors.New("temporarily unavailable")
)

type Quote struct {
	PriceKopecks int64
	ETAMinutes   int32
}

// Estimate не обращается к сети и не создаёт заказ.
func Estimate(address string, weight int64) (Quote, error) {
	if address == "" || weight <= 0 || weight > 20000 {
		return Quote{}, ErrInvalid
	}
	if address == "Недоступно" {
		return Quote{}, ErrUnavailable
	}
	var base int64
	var eta int32
	switch address {
	case "Центр":
		base, eta = 20000, 30
	case "Север":
		base, eta = 30000, 45
	default:
		return Quote{}, ErrOutside
	}
	return Quote{base + ((weight+999)/1000)*5000, eta}, nil
}
