// Слайд 17: ISP. Запуск из lecture: go run ./10-solid/02-isp
// Вопрос: почему расчёт суммы требует умения выпекать и доставлять?
package main

import "fmt"

func main() {
	fmt.Println("До, пицца из пиццерии:", BeforeItemTotal(ShopPizza{}, 2))
	fmt.Println("После, пицца из пиццерии:", ItemTotal(ShopPizza{}, 2))
	fmt.Println("После, цена из каталога:", ItemTotal(CatalogPizza{price: 600}, 2))

	// Раскомментировать: CatalogPizza не имеет Bake и Deliver.
	// BeforeItemTotal(CatalogPizza{price: 600}, 2)
	// Маленький интерфейс не требует удалять методы у реального типа.
}
