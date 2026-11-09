package kitchen

import "errors"

type Order struct {
	ID       int
	Pizza    string
	PriceRub int
}
type BakedPizza struct {
	OrderID int
	Label   string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type Baker func(Order) BakedPizza

var ErrNotImplemented = errors.New("задание пока не выполнено")
var ErrInvalidInterval = errors.New("интервал должен быть положительным")
