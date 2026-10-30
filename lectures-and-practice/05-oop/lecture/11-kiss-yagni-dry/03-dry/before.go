package main

// Новое правило: доставка бесплатна при сумме товаров от 2000 рублей.
// Изначально порог был 2500. В корзине обновили копию правила...
func BeforeCartDelivery(itemsTotal int) int {
	if itemsTotal >= 2000 {
		return 0
	}

	return 300
}

// ...а при оплате осталась старая копия. Это намеренная ошибка примера.
func BeforeCheckoutDelivery(itemsTotal int) int {
	if itemsTotal >= 2500 {
		return 0
	}

	return 300
}
