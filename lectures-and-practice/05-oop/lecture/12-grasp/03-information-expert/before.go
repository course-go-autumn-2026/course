package main

// Общая для обоих вариантов модель цены пиццы.
type Pizza struct {
	price int
}

func (p Pizza) Price() int { return p.price }

type BeforeOrderItem struct {
	Pizza Pizza
	Count int
}

type BeforeOrder struct {
	Items []BeforeOrderItem
}

// Внешний сценарий извлекает данные заказа и сам реализует расчёт.
// Для этого он должен знать состав позиции и формулу её стоимости.
func BeforeCheckoutTotal(order BeforeOrder) int {
	total := 0
	for _, item := range order.Items {
		total += item.Pizza.Price() * item.Count
	}
	return total
}
