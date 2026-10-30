package main

// Контракт определяется потребностями расчёта стоимости.
type Priced interface {
	Price() int
}

func ItemTotal(pizza Priced, count int) int {
	return pizza.Price() * count
}

// Значение из каталога умеет сообщать цену. Выпекать или доставлять
// его не требуется, поэтому соответствующих методов нет.
type CatalogPizza struct {
	price int
}

func (p CatalogPizza) Price() int { return p.price }
