package stock

import "errors"

type Ingredient struct {
	ID            int
	Name          string
	RequiredGrams int
}
type StockStatus struct {
	IngredientID int
	Description  string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type StockChecker func(Ingredient) StockStatus

var ErrNotImplemented = errors.New("задание пока не выполнено")
var (
	ErrInvalidTimeout = errors.New("таймаут должен быть положительным")
	ErrTimeout        = errors.New("истекло время ожидания")
	ErrClosed         = errors.New("канал закрыт")
)
