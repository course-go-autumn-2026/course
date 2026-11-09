package notifications

import "errors"

type Notification struct {
	ID      int
	Channel string
	OrderID int
}
type DeliveryResult struct {
	MessageID int
	Status    string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type Sender func(Notification) DeliveryResult

var ErrNotImplemented = errors.New("задание пока не выполнено")
var ErrInvalidInterval = errors.New("интервал должен быть положительным")
