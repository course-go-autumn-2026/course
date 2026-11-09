package receipts

import "errors"

type ReceiptRequest struct {
	ID        int
	Customer  string
	AmountRub int
}
type Receipt struct {
	OrderID int
	Text    string
}

// Callback безопасен при параллельных вызовах, завершается и не паникует.
type ReceiptPrinter func(ReceiptRequest) Receipt

var ErrNotImplemented = errors.New("задание пока не выполнено")
var ErrInvalidInterval = errors.New("интервал должен быть положительным")
