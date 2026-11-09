package ovens

import "errors"

type OvenTask struct {
	ID            int
	Model         string
	TargetCelsius int
}
type OvenStatus struct {
	OvenID      int
	Description string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type OvenChecker func(OvenTask) OvenStatus

var ErrNotImplemented = errors.New("задание пока не выполнено")
var (
	ErrInvalidTimeout = errors.New("таймаут должен быть положительным")
	ErrTimeout        = errors.New("истекло время ожидания")
	ErrClosed         = errors.New("канал закрыт")
)
