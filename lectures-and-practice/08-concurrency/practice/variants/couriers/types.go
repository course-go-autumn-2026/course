package couriers

import "errors"

type CourierTask struct {
	ID             int
	Name           string
	DistanceMeters int
}
type CourierCard struct {
	CourierID   int
	Description string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type CardBuilder func(CourierTask) CourierCard

var ErrNotImplemented = errors.New("задание пока не выполнено")
var (
	ErrInvalidTimeout = errors.New("таймаут должен быть положительным")
	ErrTimeout        = errors.New("истекло время ожидания")
	ErrClosed         = errors.New("канал закрыт")
)
