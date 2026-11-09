//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: у правил маршрутизации нет явной остановки и приоритета.
// Критичная заявка ошибочно уйдёт в отдел оплаты. Цепочка остановится на первом решении.
// При фиксированных двух правилах исправить if тоже достаточно.

package main

import "fmt"

type Ticket struct {
	Subject  string
	Category string
	Critical bool
}

func Route(ticket Ticket) string {
	team := "общая поддержка"
	if ticket.Critical {
		team = "дежурный инженер"
	}
	// Независимые проверки перезаписывают уже принятое решение.
	if ticket.Category == "billing" {
		team = "отдел оплаты"
	}
	return team
}

func main() {
	tickets := []Ticket{
		{Subject: "Не работает оплата у всех", Category: "billing", Critical: true},
		{Subject: "Нужен чек", Category: "billing"},
		{Subject: "Как сменить имя?", Category: "account"},
	}
	for _, ticket := range tickets {
		fmt.Printf("%s -> %s\n", ticket.Subject, Route(ticket))
	}
}
