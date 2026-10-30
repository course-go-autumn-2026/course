// Слайд 33: Information Expert.
// Запуск из lecture: go run ./12-grasp/03-information-expert
// Вопрос: у кого уже есть данные для расчёта позиции и всего заказа?
package main

import "fmt"

func main() {
	margherita, pepperoni := Pizza{price: 600}, Pizza{price: 750}
	before := BeforeOrder{Items: []BeforeOrderItem{
		{Pizza: margherita, Count: 2},
		{Pizza: pepperoni, Count: 1},
	}}
	fmt.Println("До, расчёт во внешнем сценарии:", BeforeCheckoutTotal(before))

	first := OrderItem{pizza: margherita, count: 2}
	second := OrderItem{pizza: pepperoni, count: 1}
	order := Order{items: []OrderItem{first, second}}
	fmt.Println("Первая позиция:", first.Subtotal())
	fmt.Println("Вторая позиция:", second.Subtotal())
	fmt.Println("После, расчёт поручен заказу:", CheckoutTotal(order))
	fmt.Println("Пустой заказ:", CheckoutTotal(Order{}))
}
