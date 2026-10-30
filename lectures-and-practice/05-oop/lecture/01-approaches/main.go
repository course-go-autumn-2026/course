// Запуск: go run ./01-approaches
// Вопрос: после какого добавления исходный заказ останется пустым?
package main

import (
	"errors"
	"fmt"
)

// Для краткости цены считаем корректными целыми рублями, без переполнения.
type Item struct {
	Price int
	Count int
}

type OrderData struct {
	Items []Item
}

// Процедурный вариант: операция изменяет переданные данные.
func AddItem(order *OrderData, item Item) error {
	if item.Count <= 0 {
		return errors.New("count must be positive")
	}
	order.Items = append(order.Items, item)
	return nil
}

// Функциональный вариант: вычисляем новый заказ, не меняя исходный.
// Копии структуры недостаточно: срезы могли бы разделять один массив.
// Здесь Item содержит только int, поэтому копии элементов достаточно.
func WithItem(order OrderData, item Item) (OrderData, error) {
	if item.Count <= 0 {
		return OrderData{}, errors.New("count must be positive")
	}
	items := make([]Item, len(order.Items)+1)
	copy(items, order.Items)
	items[len(order.Items)] = item
	return OrderData{Items: items}, nil
}

// Чистый расчёт можно использовать в любом из трёх подходов.
func Total(items []Item) int {
	total := 0
	for _, item := range items {
		total += item.Price * item.Count
	}
	return total
}

// Объектный вариант: типу поручены состояние и проверка его изменений.
// items скрыт от других пакетов, но доступен коду этого же пакета.
type Order struct {
	items []Item
}

func (o *Order) Add(item Item) error {
	if item.Count <= 0 {
		return errors.New("count must be positive")
	}
	o.items = append(o.items, item)
	return nil
}

func (o Order) Total() int { return Total(o.items) }

func main() {
	item := Item{Price: 600, Count: 2}
	var procedural OrderData
	if err := AddItem(&procedural, item); err != nil {
		panic(err)
	}
	fmt.Println("procedural:", Total(procedural.Items))

	original := OrderData{}
	updated, err := WithItem(original, item)
	if err != nil {
		panic(err)
	}
	fmt.Println("functional original:", Total(original.Items))
	fmt.Println("functional updated:", Total(updated.Items))

	// Проверим, что обновлённый заказ не разделяет массив с предыдущим.
	next, err := WithItem(updated, Item{Price: 750, Count: 1})
	if err != nil {
		panic(err)
	}
	next.Items[0].Count = 99
	fmt.Println("previous order unchanged:", Total(updated.Items))

	var object Order
	if err := object.Add(item); err != nil {
		panic(err)
	}
	err = object.Add(Item{Price: 600, Count: -1})
	fmt.Println("object:", object.Total())
	fmt.Println("invalid count rejected:", err != nil)
}
