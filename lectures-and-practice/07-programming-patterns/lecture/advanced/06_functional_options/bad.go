//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: позиционные настройки плохо читаются и не отличают пропуск от пустого значения.
// Functional Options дадут имена настройкам и отдельную проверку явно заданного значения.

package main

import (
	"errors"
	"fmt"
	"strings"
)

type Greeter struct {
	name      string
	greeting  string
	uppercase bool
}

func NewGreeter(name, greeting string, uppercase bool) (*Greeter, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("имя не должно быть пустым")
	}
	// Пустая строка одновременно означает default и ошибочное значение.
	if greeting == "" {
		greeting = "Привет"
	}
	return &Greeter{name: name, greeting: greeting, uppercase: uppercase}, nil
}

func (g *Greeter) Greet() string {
	message := g.greeting + ", " + g.name + "!"
	if g.uppercase {
		return strings.ToUpper(message)
	}
	return message
}

func main() {
	g, err := NewGreeter("Анна", "Добрый день", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(g.Greet())

	_, err = NewGreeter("Анна", "", false)
	fmt.Println("Неверная настройка:", err) // Явно пустое приветствие не отклонено.
}
