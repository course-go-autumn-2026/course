// Слайды 21–22: YAGNI. Запуск из lecture: go run ./11-kiss-yagni-dry/02-yagni
// Вопрос: какие требования сейчас оправдывают реестр способов доставки?
package main

import "fmt"

func main() {
	const pizzaTotal = 1950

	registry := NewDeliveryRegistry()
	registry.Register("courier", func() BeforeDelivery { return BeforeCourier{} })
	delivery, err := registry.Create("courier")
	if err != nil {
		panic(err)
	}
	fmt.Println("До, итог заказа:", pizzaTotal+delivery.Fee())
	fmt.Println("После, итог заказа:", OrderTotal(pizzaTotal))

	// Дополнительная настройка создаёт дополнительный сценарий ошибки.
	_, err = registry.Create("drone")
	fmt.Println("Ошибка лишней настройки:", err)
}
