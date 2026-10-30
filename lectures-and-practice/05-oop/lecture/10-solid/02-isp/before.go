package main

import "fmt"

type Pizza interface {
	Price() int
	Bake() error
	Deliver(address string) error
}

// Для расчёта требуется только Price, но параметр требует все три метода.
func BeforeItemTotal(pizza Pizza, count int) int {
	return pizza.Price() * count
}

type ShopPizza struct{}

func (ShopPizza) Price() int { return 600 }
func (ShopPizza) Bake() error {
	fmt.Println("Выпекаем пиццу")
	return nil
}
func (ShopPizza) Deliver(address string) error {
	fmt.Println("Доставляем по адресу:", address)
	return nil
}
