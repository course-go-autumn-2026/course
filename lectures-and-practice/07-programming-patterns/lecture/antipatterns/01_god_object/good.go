//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: расчёт зарплат и управление принтером связаны одним объектом.
// Решение: разделить ответственности на Payroll и Printer, используя композицию.

package main

import "fmt"

// После: расчёт зарплат и работа принтера имеют отдельных владельцев.
// Дополнительные интерфейсы для этого примера не нужны.
type Payroll struct {
	Salaries []int
}

func (p Payroll) Total() int {
	total := 0
	for _, salary := range p.Salaries {
		total += salary
	}
	return total
}

type Printer struct {
	Paper int
}

func (p *Printer) Print(text string) {
	if p.Paper == 0 {
		fmt.Println("Нет бумаги")
		return
	}
	p.Paper--
	fmt.Println(text)
}

func main() {
	payroll := Payroll{Salaries: []int{100, 150}}
	printer := Printer{Paper: 1}
	printer.Print(fmt.Sprintf("Фонд зарплаты: %d", payroll.Total()))
}
