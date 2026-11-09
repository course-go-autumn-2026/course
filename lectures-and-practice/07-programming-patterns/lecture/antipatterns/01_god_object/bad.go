//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: расчёт зарплат и управление принтером связаны одним объектом.
// Решение: разделить ответственности на Payroll и Printer, используя композицию.

package main

import "fmt"

// Уменьшенная модель God Object: один объект знает и про зарплаты,
// и про расход бумаги в принтере. В большом приложении такие роли растут независимо.
type OfficeBefore struct {
	salaries []int
	paper    int
}

func (o *OfficeBefore) PrintPayroll() {
	total := 0
	for _, salary := range o.salaries {
		total += salary
	}
	if o.paper == 0 {
		fmt.Println("Нет бумаги")
		return
	}
	o.paper--
	fmt.Printf("Фонд зарплаты: %d\n", total)
}

func main() {
	office := OfficeBefore{salaries: []int{100, 150}, paper: 1}
	office.PrintPayroll()
}
