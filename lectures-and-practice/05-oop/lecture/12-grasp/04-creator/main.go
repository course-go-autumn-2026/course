// Слайд 34: Creator. Запуск из lecture: go run ./12-grasp/04-creator
// Вопрос: кто должен знать, как создать позицию и включить её в заказ?
package main

import "fmt"

func main() {
	before := BeforeOrder{}
	order := Order{}
	for _, input := range []struct {
		pizza Pizza
		count int
	}{{Pizza{price: 600}, 2}, {Pizza{price: 750}, 1}} {
		if err := BeforeCheckoutAdd(&before, input.pizza, input.count); err != nil {
			panic(err)
		}
		if err := order.AddPizza(input.pizza, input.count); err != nil {
			panic(err)
		}
	}
	fmt.Println("До, позиции создаёт внешний сценарий:", before.Total())
	fmt.Println("После, позиции создаёт заказ:", order.Total())

	fmt.Println("До, некорректное количество:", BeforeCheckoutAdd(&before, Pizza{price: 600}, -1))
	fmt.Println("После, некорректное количество:", order.AddPizza(Pizza{price: 600}, -1))
	fmt.Println("Суммы после отказа:", before.Total(), order.Total())
}
