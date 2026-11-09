//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import (
	"fmt"
	"strings"
)

type Text interface {
	Render() string
}

type PlainText string

func (t PlainText) Render() string {
	return string(t)
}

// Каждый декоратор принимает Text и сам остаётся Text.
type Uppercase struct {
	Inner Text
}

func (u Uppercase) Render() string {
	return strings.ToUpper(u.Inner.Render())
}

type WithLabel struct {
	Inner Text
	Label string
}

func (w WithLabel) Render() string {
	return w.Label + ": " + w.Inner.Render()
}

func main() {
	text := PlainText("привет")
	fmt.Println(text.Render())

	// Порядок обёрток влияет на результат.
	fmt.Println(WithLabel{Inner: Uppercase{Inner: text}, Label: "Сообщение"}.Render())
	fmt.Println(Uppercase{Inner: WithLabel{Inner: text, Label: "Сообщение"}}.Render())
}
