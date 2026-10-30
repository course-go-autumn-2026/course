// Слайд 18: OCP. Запуск из lecture: go run ./10-solid/03-ocp
// Вопрос: какой код изменится при появлении экспресс-доставки?
package main

import "fmt"

func main() {
	const pizzaTotal, distanceKm = 1950, 5
	fmt.Println("До, самовывоз:", BeforeOrderTotal(pizzaTotal, distanceKm, true))
	fmt.Println("До, курьер:", BeforeOrderTotal(pizzaTotal, distanceKm, false))
	fmt.Println("После, самовывоз:", OrderTotal(pizzaTotal, distanceKm, Pickup{}))
	fmt.Println("После, курьер:", OrderTotal(pizzaTotal, distanceKm, Courier{}))
	fmt.Println("Новая экспресс-доставка:", OrderTotal(pizzaTotal, distanceKm, Express{}))
}
