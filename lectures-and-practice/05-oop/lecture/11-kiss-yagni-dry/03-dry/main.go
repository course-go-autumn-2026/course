// Слайд 21: DRY. Запуск из lecture: go run ./11-kiss-yagni-dry/03-dry
// Вопрос: почему корзина и оплата показывают разную стоимость доставки?
package main

import "fmt"

func main() {
	// Проверяем границу нового правила и прежний порог.
	for _, total := range []int{1999, 2000, 2500} {
		fmt.Printf("Товары на %d руб.\n", total)
		fmt.Printf("До: корзина = %d, оплата = %d\n",
			BeforeCartDelivery(total), BeforeCheckoutDelivery(total))
		fmt.Printf("После: корзина = %d, оплата = %d\n",
			CartDelivery(total), CheckoutDelivery(total))
	}
}
