//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import "fmt"

type Ticket struct {
	Subject  string
	Category string
	Critical bool
}

// Второй результат сообщает, взял ли обработчик заявку.
type Handler func(Ticket) (team string, handled bool)

func CriticalSupport(ticket Ticket) (string, bool) {
	if ticket.Critical {
		return "дежурный инженер", true
	}
	return "", false
}

func BillingSupport(ticket Ticket) (string, bool) {
	if ticket.Category == "billing" {
		return "отдел оплаты", true
	}
	return "", false
}

func GeneralSupport(Ticket) (string, bool) {
	return "общая поддержка", true
}

func Route(ticket Ticket, chain ...Handler) (string, error) {
	for _, handle := range chain {
		if team, handled := handle(ticket); handled {
			return team, nil // Первый подходящий обработчик завершает цепочку.
		}
	}
	return "", fmt.Errorf("для заявки %q нет обработчика", ticket.Subject)
}

func main() {
	// Критичность важнее категории. Универсальный обработчик идёт последним.
	chain := []Handler{CriticalSupport, BillingSupport, GeneralSupport}
	tickets := []Ticket{
		{Subject: "Не работает оплата у всех", Category: "billing", Critical: true},
		{Subject: "Нужен чек", Category: "billing"},
		{Subject: "Как сменить имя?", Category: "account"},
	}
	for _, ticket := range tickets {
		team, err := Route(ticket, chain...)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%s -> %s\n", ticket.Subject, team)
	}
}
