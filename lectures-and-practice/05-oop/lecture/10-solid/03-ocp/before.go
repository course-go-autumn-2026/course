package main

// Новый способ доставки потребует изменить параметры или ветвления
// внутри этой функции. Для двух постоянных вариантов это может быть нормально.
func BeforeOrderTotal(pizzaTotal, distanceKm int, pickup bool) int {
	if pickup {
		return pizzaTotal
	}

	return pizzaTotal + 200 + 20*distanceKm
}
