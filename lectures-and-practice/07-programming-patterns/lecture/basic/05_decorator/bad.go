//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: центральная функция перечисляет комбинации оформления.
// Новая обёртка умножает число режимов. Decorator позволяет составлять их по отдельности.

package main

import (
	"fmt"
	"strings"
)

func Render(text string, mode string) string {
	// Для каждой комбинации оформления создаётся отдельная ветка.
	switch mode {
	case "plain":
		return text
	case "uppercase_with_label":
		return "Сообщение: " + strings.ToUpper(text)
	case "label_then_uppercase":
		return strings.ToUpper("Сообщение: " + text)
	default:
		return text
	}
}

func main() {
	fmt.Println(Render("привет", "plain"))
	fmt.Println(Render("привет", "uppercase_with_label"))
	fmt.Println(Render("привет", "label_then_uppercase"))
}
