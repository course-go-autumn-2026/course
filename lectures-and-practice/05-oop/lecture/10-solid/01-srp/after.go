package main

import "fmt"

// Заказ отвечает за данные и расчёт, а форматирование вынесено отдельно.
type Order struct {
	prices []int
}

func (o Order) Total() int {
	total := 0
	for _, price := range o.prices {
		total += price
	}
	return total
}

// Для этого разделения не нужны интерфейсы или сложная иерархия типов.
func ReceiptRU(total int) string {
	return fmt.Sprintf("Итого: %d руб.", total)
}

func ReceiptEN(total int) string {
	return fmt.Sprintf("Total: %d RUB", total)
}
