//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: пошаговая сборка URL вручную не проверяет обязательные данные.
// Builder соберёт настройки и проверит результат; кодирование поручим net/url.

package main

import "fmt"

func main() {
	host, path, query, page := "example.com", "/search", "go patterns", "1"
	// Порядок сборки и кодирование параметров оставлены вызывающему коду.
	address := "https://" + host
	address += path
	address += "?q=" + query
	address += "&page=" + page
	fmt.Println(address) // Пробел не закодирован.

	host = ""
	address = "https://" + host + path
	fmt.Println("Сборка без хоста:", address) // Ошибка создания не обнаружена.
}
