package main

// Предусмотренная точка расширения - правило расчёта доставки.
// В примере цены и расстояния неотрицательные, переполнение исключено.
type Delivery interface {
	Fee(distanceKm int) int
}

type Pickup struct{}

func (Pickup) Fee(int) int { return 0 }

type Courier struct{}

func (Courier) Fee(distanceKm int) int { return 200 + 20*distanceKm }

func OrderTotal(pizzaTotal, distanceKm int, delivery Delivery) int {
	return pizzaTotal + delivery.Fee(distanceKm)
}

// Новый вариант добавляется без изменения OrderTotal.
// При выборе варианта всё равно нужно настроить его в вызывающем коде.
type Express struct{}

func (Express) Fee(distanceKm int) int { return 500 + 30*distanceKm }
