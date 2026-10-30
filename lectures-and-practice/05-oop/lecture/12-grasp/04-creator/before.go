package main

import "fmt"

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

func (order BeforeOrder) Total() int {
	total := 0
	for _, item := range order.Items {
		total += item.Pizza.Price() * item.Count
	}
	return total
}

// Внешний сценарий знает, как проверить, создать и присоединить позицию.
// Другим сценариям тоже придётся знать эту процедуру или зависеть от неё.
func BeforeCheckoutAdd(order *BeforeOrder, pizza Pizza, count int) error {
	if count <= 0 {
		return fmt.Errorf("количество должно быть положительным")
	}
	item := BeforeOrderItem{Pizza: pizza, Count: count}
	order.Items = append(order.Items, item)

	return nil
}
