package routes

import "errors"

type Delivery struct {
	ID             int
	Address        string
	DistanceMeters int
}
type Route struct {
	OrderID     int
	Description string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type RouteEstimator func(Delivery) Route

var ErrNotImplemented = errors.New("задание пока не выполнено")
var (
	ErrInvalidTimeout = errors.New("таймаут должен быть положительным")
	ErrTimeout        = errors.New("истекло время ожидания")
	ErrClosed         = errors.New("канал закрыт")
)
