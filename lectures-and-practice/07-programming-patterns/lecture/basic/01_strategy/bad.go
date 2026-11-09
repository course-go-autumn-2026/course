//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: ради другого заголовка продублирован весь сценарий отчёта.
// Добавить строку отчёта придётся в каждую функцию. Strategy оставляет один сценарий.

package main

import (
	"fmt"
	"strings"
)

func PrintUpperReport(title string) {
	fmt.Println(strings.ToUpper(title))
	fmt.Println("Продано книг: 12")
}

func PrintFramedReport(title string) {
	fmt.Println("*** " + title + " ***")
	fmt.Println("Продано книг: 12")
}

func main() {
	PrintUpperReport("Отчёт магазина")
	fmt.Println()
	PrintFramedReport("Отчёт магазина")
}
