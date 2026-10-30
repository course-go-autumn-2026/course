// Запуск из lecture: go run ./01-approaches/procedural
// Вопрос: кто хранит состояние заказа и какие функции его изменяют?
package main

import "fmt"

type Item struct {
	Price int // Целые рубли. Для примера не рассматриваем переполнение.
	Count int
}

type Order struct {
	Items []Item
}

// AddItem изменяет переданный заказ. Проверки нужны в любом подходе.
func AddItem(order *Order, item Item) error {
	if item.Count <= 0 {
		return fmt.Errorf("количество должно быть положительным")
	}
	if item.Price < 0 {
		return fmt.Errorf("цена не может быть отрицательной")
	}
	order.Items = append(order.Items, item)
	return nil
}

// Процедурная программа тоже может использовать чистые функции.
func Total(order Order) int {
	total := 0
	for _, item := range order.Items {
		total += item.Price * item.Count
	}
	return total
}

func main() {
	order := Order{}
	fmt.Println("Пустой заказ:", Total(order))

	if err := AddItem(&order, Item{Price: 600, Count: 2}); err != nil {
		panic(err) // В учебном сценарии корректные данные обязаны пройти.
	}
	fmt.Println("После первого добавления:", Total(order))

	if err := AddItem(&order, Item{Price: 750, Count: 1}); err != nil {
		panic(err)
	}
	fmt.Println("После второго добавления:", Total(order))

	// Обе процедуры работают с одним и тем же заказом по указателю.
	// Items экспортирован: вызывающий код может обойти проверки AddItem.
	err := AddItem(&order, Item{Price: 600, Count: -1})
	fmt.Println("Некорректное количество:", err)
	err = AddItem(&order, Item{Price: -100, Count: 1})
	fmt.Println("Некорректная цена:", err)
	fmt.Println("После отклонённых операций:", Total(order))
}
