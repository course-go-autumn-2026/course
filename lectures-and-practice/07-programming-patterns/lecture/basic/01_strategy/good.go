//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import (
	"fmt"
	"strings"
)

// Strategy: способ оформления заголовка передаётся как функция.
// Сценарий печати отчёта не меняется при замене алгоритма.
type Formatter func(string) string

func PrintReport(title string, format Formatter) {
	fmt.Println(format(title))
	fmt.Println("Продано книг: 12")
}

func main() {
	PrintReport("Отчёт магазина", strings.ToUpper)
	fmt.Println()
	PrintReport("Отчёт магазина", func(title string) string {
		return "*** " + title + " ***"
	})
}
