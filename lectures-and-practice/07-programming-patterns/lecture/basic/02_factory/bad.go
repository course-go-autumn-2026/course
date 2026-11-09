//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: каждый сценарий сам выбирает и создаёт отправителя.
// Новый канал потребует правок во всех сценариях. Фабрика соберёт выбор в одном месте.
// Отправка учебная: оба файла только печатают в консоль.

package main

import "fmt"

type EmailNotifier struct{ Address string }

func (n EmailNotifier) Notify(message string) {
	fmt.Printf("Email для %s: %s\n", n.Address, message)
}

type SMSNotifier struct{ Phone string }

func (n SMSNotifier) Notify(message string) {
	fmt.Printf("SMS для %s: %s\n", n.Phone, message)
}

func NotifyReservation(channel, recipient string) {
	switch channel {
	case "email":
		n := EmailNotifier{Address: recipient}
		n.Notify("Книга забронирована")
	case "sms":
		n := SMSNotifier{Phone: recipient}
		n.Notify("Книга забронирована")
	default:
		fmt.Println("Ошибка: неизвестный канал:", channel)
	}
}

func NotifyReminder(channel, recipient string) {
	// Та же логика выбора и создания повторена в другом сценарии.
	switch channel {
	case "email":
		n := EmailNotifier{Address: recipient}
		n.Notify("Верните книгу завтра")
	case "sms":
		n := SMSNotifier{Phone: recipient}
		n.Notify("Верните книгу завтра")
	default:
		fmt.Println("Ошибка: неизвестный канал:", channel)
	}
}

func main() {
	for _, channel := range []string{"email", "sms", "fax"} {
		NotifyReservation(channel, "Анны")
		NotifyReminder(channel, "Анны")
	}
}
