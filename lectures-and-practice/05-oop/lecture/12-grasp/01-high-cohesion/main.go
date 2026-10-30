// Слайд 30: High Cohesion. Запуск из lecture: go run ./12-grasp/01-high-cohesion
// Вопрос: какое отношение отчёт о выручке имеет к составу конкретной пиццы?
package main

import "fmt"

func main() {
	before := MixedPizza{}
	before.AddTopping("грибы")
	fmt.Println("До, есть грибы:", before.HasTopping("грибы"))
	fmt.Println("До:", before.SalesReport(1950))

	pizza := Pizza{}
	pizza.AddTopping("грибы")
	fmt.Println("После, есть грибы:", pizza.HasTopping("грибы"))
	fmt.Println("После, есть ананас:", pizza.HasTopping("ананас"))
	fmt.Println("После:", SalesReport(1950))
}
