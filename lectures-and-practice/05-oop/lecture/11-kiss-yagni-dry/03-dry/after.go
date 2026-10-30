package main

// Единое представление одного бизнес-правила для обоих потребителей.
func DeliveryFee(itemsTotal int) int {
	const freeDeliveryFrom = 2000
	const regularDeliveryFee = 300

	if itemsTotal >= freeDeliveryFrom {
		return 0
	}

	return regularDeliveryFee
}

// Две точки использования правила. Похожие строки вызова допустимы:
// само условие бесплатной доставки определено в одном месте.
func CartDelivery(itemsTotal int) int {
	return DeliveryFee(itemsTotal)
}

func CheckoutDelivery(itemsTotal int) int {
	return DeliveryFee(itemsTotal)
}
