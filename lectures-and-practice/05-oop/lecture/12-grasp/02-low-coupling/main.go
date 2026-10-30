// Слайд 31: Low Coupling. Запуск из lecture: go run ./12-grasp/02-low-coupling
// Вопрос: почему изменение состава пиццы требует менять код чека?
package main

import "fmt"

func main() {
	dough, cheese, sauce := Ingredient{Price: 100}, Ingredient{Price: 200}, Ingredient{Price: 50}
	topping := Ingredient{Price: 100}

	before := DetailedPizza{Dough: dough, Cheese: cheese, Sauce: sauce}
	fmt.Println("До, три ингредиента:", BeforeReceiptTotal(before))
	before.Topping = topping
	fmt.Println("До, добавили начинку:", BeforeReceiptTotal(before))

	fmt.Println("После, три ингредиента:", ReceiptTotal(NewPizza(dough, cheese, sauce)))
	fmt.Println("После, добавили начинку:", ReceiptTotal(NewPizza(dough, cheese, sauce, topping)))
}
