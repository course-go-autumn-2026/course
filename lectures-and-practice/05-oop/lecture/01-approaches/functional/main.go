// Запуск из lecture: go run ./01-approaches/functional
// Вопрос: какие версии заказа останутся доступны после двух добавлений?
package main

import "fmt"

type Item struct {
	Price int // Целые рубли. Для примера не рассматриваем переполнение.
	Count int
}

type Order struct {
	Items []Item
}

// WithItem вычисляет новый заказ, не изменяя исходный.
// При ошибке возвращается нулевой результат: его использовать не нужно.
func WithItem(order Order, item Item) (Order, error) {
	if item.Count <= 0 {
		return Order{}, fmt.Errorf("количество должно быть положительным")
	}
	if item.Price < 0 {
		return Order{}, fmt.Errorf("цена не может быть отрицательной")
	}

	// Копия Order сама по себе не копирует массив среза Items.
	// Новый массив исключает совместное владение элементами двух заказов.
	// У Item только поля int, поэтому копии элементов достаточно.
	items := make([]Item, len(order.Items)+1)
	copy(items, order.Items)
	items[len(order.Items)] = item
	return Order{Items: items}, nil
}

// Локальный цикл и изменение локальной переменной не мешают чистоте функции.
func Total(order Order) int {
	total := 0
	for _, item := range order.Items {
		total += item.Price * item.Count
	}
	return total
}

func main() {
	empty := Order{}

	first, err := WithItem(empty, Item{Price: 600, Count: 2})
	if err != nil {
		panic(err)
	}

	second, err := WithItem(first, Item{Price: 750, Count: 1})
	if err != nil {
		panic(err)
	}

	// Выводим ВСЕ версии уже после обоих добавлений.
	// Состояние empty и first сохранилось.
	fmt.Println("Исходный заказ после двух добавлений:", Total(empty))
	fmt.Println("Первая версия после двух добавлений:", Total(first))
	fmt.Println("Вторая версия:", Total(second))

	_, err = WithItem(second, Item{Price: 600, Count: -1})
	fmt.Println("Некорректное количество:", err)
	_, err = WithItem(second, Item{Price: -100, Count: 1})
	fmt.Println("Некорректная цена:", err)
	fmt.Println("После отклонённых операций:", Total(second))
	// fmt.Println имеет побочный эффект. Вывод отделён от вычислений.
	// Go не запрещает менять Items напрямую: неизменяемость здесь - соглашение.
}
