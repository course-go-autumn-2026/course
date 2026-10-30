// Запуск из lecture: go run ./01-approaches/object
// Вопрос: через какие операции внешний пакет может изменить заказ?
package main

import "fmt"

type Item struct {
	Price int // Целые рубли. Для примера не рассматриваем переполнение.
	Count int
}

// Order объединяет состояние заказа и операции над ним.
// items скрыт от других пакетов. Код этого же пакета имеет к нему доступ.
type Order struct {
	items []Item
}

// Add сохраняет инварианты: Count > 0 и Price >= 0 у каждой позиции.
// Проверки происходят до изменения состояния, поэтому ошибка его не меняет.
func (o *Order) Add(item Item) error {
	if item.Count <= 0 {
		return fmt.Errorf("количество должно быть положительным")
	}
	if item.Price < 0 {
		return fmt.Errorf("цена не может быть отрицательной")
	}
	o.items = append(o.items, item)
	return nil
}

// Метод может только вычислять результат, не меняя объект.
func (o *Order) Total() int {
	total := 0
	for _, item := range o.items {
		total += item.Price * item.Count
	}
	return total
}

func main() {
	order := Order{} // Нулевое значение уже является корректным пустым заказом.
	fmt.Println("Пустой заказ:", order.Total())

	if err := order.Add(Item{Price: 600, Count: 2}); err != nil {
		panic(err)
	}
	fmt.Println("После первого добавления:", order.Total())

	if err := order.Add(Item{Price: 750, Count: 1}); err != nil {
		panic(err)
	}
	fmt.Println("После второго добавления:", order.Total())

	err := order.Add(Item{Price: 600, Count: -1})
	fmt.Println("Некорректное количество:", err)
	err = order.Add(Item{Price: -100, Count: 1})
	fmt.Println("Некорректная цена:", err)
	fmt.Println("После отклонённых операций:", order.Total())
}
