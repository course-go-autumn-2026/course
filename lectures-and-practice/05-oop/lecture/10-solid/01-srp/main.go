// Слайд 16: SRP. Запуск из lecture: go run ./10-solid/01-srp
// Вопрос: почему новый язык чека должен затрагивать расчёт заказа?
package main

import "fmt"

func main() {
	before := MixedOrder{prices: []int{600, 600, 750}}
	fmt.Println("До:", before.Receipt())

	order := Order{prices: []int{600, 600, 750}}
	fmt.Println("После, русский:", ReceiptRU(order.Total()))
	fmt.Println("После, английский:", ReceiptEN(order.Total()))
	fmt.Println("Сумма заказа:", order.Total())
}
