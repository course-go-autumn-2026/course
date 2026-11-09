package promos

import "errors"

type PromoRequest struct {
	ID      int
	Code    string
	BaseRub int
}
type Quote struct {
	OrderID     int
	Description string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type QuoteBuilder func(PromoRequest) Quote

var ErrNotImplemented = errors.New("задание пока не выполнено")
var ErrInvalidInterval = errors.New("интервал должен быть положительным")
