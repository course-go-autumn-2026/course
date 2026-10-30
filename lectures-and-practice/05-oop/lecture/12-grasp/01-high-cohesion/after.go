package main

import "fmt"

// Оба метода работают с составом пиццы. Несколько методов у типа - нормально.
type Pizza struct {
	toppings []string
}

func (p *Pizza) AddTopping(name string) {
	p.toppings = append(p.toppings, name)
}

func (p Pizza) HasTopping(name string) bool {
	for _, topping := range p.toppings {
		if topping == name {
			return true
		}
	}
	return false
}

// Отчёт получает только необходимые ему данные. Отдельный тип не требуется.
func SalesReport(revenue int) string {
	return fmt.Sprintf("Выручка за день: %d руб.", revenue)
}
