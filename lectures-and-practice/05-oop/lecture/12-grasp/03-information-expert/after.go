package main

type OrderItem struct {
	pizza Pizza
	count int
}

// У позиции есть цена и количество: ей поручаем расчёт её стоимости.
func (item OrderItem) Subtotal() int {
	return item.pizza.Price() * item.count
}

type Order struct {
	items []OrderItem
}

// У заказа есть позиции: ему поручаем сумму, делегируя расчёт каждой позиции.
func (order Order) Total() int {
	total := 0
	for _, item := range order.items {
		total += item.Subtotal()
	}
	return total
}

func CheckoutTotal(order Order) int { return order.Total() }
