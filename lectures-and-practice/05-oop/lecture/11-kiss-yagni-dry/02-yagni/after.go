package main

// Решаем текущее требование. Расчёт выделен, чтобы его легко найти и изменить,
// когда появится обоснованная потребность в других способах доставки.
func OrderTotal(pizzaTotal int) int {
	const deliveryFee = 300
	return pizzaTotal + deliveryFee
}
