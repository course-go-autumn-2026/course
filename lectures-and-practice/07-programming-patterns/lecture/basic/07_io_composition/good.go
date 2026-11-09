//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Функции достаточно io.Writer: можно передать буфер, файл или stdout.
func WriteList(w io.Writer, items []string) error {
	for _, item := range items {
		if _, err := fmt.Fprintln(w, item); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var buffer bytes.Buffer
	if err := WriteList(&buffer, []string{"milk", "bread", "tea"}); err != nil {
		fmt.Println("Ошибка записи:", err)
		return
	}
	fmt.Println("Весь список:")
	fmt.Print(buffer.String())

	// Композиция Reader: LimitReader оборачивает буфер и ограничивает чтение.
	// Лимит измеряется в байтах. Здесь ASCII, поэтому символ занимает один байт.
	fmt.Println("Первые 11 байт:")
	if _, err := io.Copy(os.Stdout, io.LimitReader(&buffer, 11)); err != nil {
		fmt.Println("Ошибка чтения:", err)
	}
}
