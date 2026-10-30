package main

import "fmt"

// Условие задачи: расчёт цены меняется по бизнес-правилам, а язык и формат
// чека - по требованиям к представлению. Эти причины изменения независимы.
type MixedOrder struct {
	prices []int
}

func (o MixedOrder) Total() int {
	total := 0
	for _, price := range o.prices {
		total += price
	}
	return total
}

// Добавление другого языка чека потребует править тип заказа.
func (o MixedOrder) Receipt() string {
	return fmt.Sprintf("Итого: %d руб.", o.Total())
}
