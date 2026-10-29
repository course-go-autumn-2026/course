package pizzeria

import "errors"

// Все цены в этой практике — целые рубли.
type Pizza interface {
	Name() string
	Price() int
}

type Margherita struct{}
type Pepperoni struct{}

type WithCheese struct{ Base Pizza }
type WithBox struct{ Base Pizza }

// Стратегия — обычная функция: цена пиццы -> итог с доставкой.
type DeliveryPrice func(base int) int

var ErrUnknownPizza = errors.New("unknown pizza")
var ErrUnknownDelivery = errors.New("unknown delivery")
var ErrNotRefactored = errors.New("not refactored")
