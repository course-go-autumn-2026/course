package packing

import "errors"

type PackRequest struct {
	ID       int
	Size     string
	PriceRub int
}
type PackedOrder struct {
	OrderID     int
	Description string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type Packer func(PackRequest) PackedOrder

var ErrNotImplemented = errors.New("задание пока не выполнено")
var ErrInvalidInterval = errors.New("интервал должен быть положительным")
