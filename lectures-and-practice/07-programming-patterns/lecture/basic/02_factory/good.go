//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import "fmt"

// Простая фабрика выбирает реализацию за интерфейсом.
// Это не GoF Factory Method: здесь нет иерархии создателей.
type Notifier interface {
	Notify(message string)
}

type EmailNotifier struct {
	Address string
}

func (n EmailNotifier) Notify(message string) {
	// Учебная имитация: настоящая отправка не выполняется.
	fmt.Printf("Email для %s: %s\n", n.Address, message)
}

type SMSNotifier struct {
	Phone string
}

func (n SMSNotifier) Notify(message string) {
	fmt.Printf("SMS для %s: %s\n", n.Phone, message)
}

func NewNotifier(channel, recipient string) (Notifier, error) {
	switch channel {
	case "email":
		return EmailNotifier{Address: recipient}, nil
	case "sms":
		return SMSNotifier{Phone: recipient}, nil
	default:
		return nil, fmt.Errorf("неизвестный канал: %s", channel)
	}
}

func NotifyReservation(channel, recipient string) {
	notifier, err := NewNotifier(channel, recipient)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	notifier.Notify("Книга забронирована")
}

func NotifyReminder(channel, recipient string) {
	notifier, err := NewNotifier(channel, recipient)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	notifier.Notify("Верните книгу завтра")
}

func main() {
	for _, channel := range []string{"email", "sms", "fax"} {
		NotifyReservation(channel, "Анны")
		NotifyReminder(channel, "Анны")
	}
}
