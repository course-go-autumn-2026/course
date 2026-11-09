package feedback

import "errors"

type ReviewRequest struct {
	ID      int
	Product string
	Stars   int
}
type Review struct {
	ReviewID int
	Text     string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type ReviewLoader func(ReviewRequest) Review

var ErrNotImplemented = errors.New("задание пока не выполнено")
var ErrInvalidInterval = errors.New("интервал должен быть положительным")
