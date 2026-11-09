//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

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

type Option func(*Greeter) error

func WithGreeting(greeting string) Option {
	return func(g *Greeter) error {
		if strings.TrimSpace(greeting) == "" {
			return errors.New("приветствие не должно быть пустым")
		}
		g.greeting = greeting
		return nil
	}
}

func WithUppercase() Option {
	return func(g *Greeter) error {
		g.uppercase = true
		return nil
	}
}

// Обязательное имя передаётся явно. Необязательные настройки имеют defaults.
func NewGreeter(name string, options ...Option) (*Greeter, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("имя не должно быть пустым")
	}
	g := &Greeter{name: name, greeting: "Привет"}
	for _, option := range options {
		if err := option(g); err != nil {
			return nil, err
		}
	}
	return g, nil
}

func (g *Greeter) Greet() string {
	message := g.greeting + ", " + g.name + "!"
	if g.uppercase {
		return strings.ToUpper(message)
	}
	return message
}

func main() {
	g, err := NewGreeter("Анна", WithGreeting("Добрый день"), WithUppercase())
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println(g.Greet())

	_, err = NewGreeter("Анна", WithGreeting(""))
	fmt.Println("Неверная настройка:", err)
}
