package main

import "fmt"

type orderItem struct {
	pizza Pizza
	count int
}

type Order struct {
	items []orderItem
}

// Заказ содержит позиции и получает данные для их создания.
// Поэтому ему поручены создание и включение позиции в свой состав.
func (order *Order) AddPizza(pizza Pizza, count int) error {
	if count <= 0 {
		return fmt.Errorf("количество должно быть положительным")
	}
	item := orderItem{pizza: pizza, count: count}
	order.items = append(order.items, item)
	return nil
}

func (order Order) Total() int {
	total := 0
	for _, item := range order.items {
		total += item.pizza.Price() * item.count
	}
	return total
}
