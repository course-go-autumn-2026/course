package main

import "fmt"

// Управление составом пиццы и финансовый отчёт не используют общие данные.
// Их объединение в одном типе снижает связность его обязанностей.
type MixedPizza struct {
	toppings []string
}

func (p *MixedPizza) AddTopping(name string) {
	p.toppings = append(p.toppings, name)
}

func (p MixedPizza) HasTopping(name string) bool {
	for _, topping := range p.toppings {
		if topping == name {
			return true
		}
	}
	return false
}

func (MixedPizza) SalesReport(revenue int) string {
	return fmt.Sprintf("Выручка за день: %d руб.", revenue)
}
