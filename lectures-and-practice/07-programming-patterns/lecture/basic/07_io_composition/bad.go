//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: форматирование повторяется для каждого типа назначения.
// io.Writer отделит формат от назначения, а io.LimitReader ограничит любой Reader.

package main

import (
	"bytes"
	"fmt"
	"os"
)

func WriteListToBuffer(buffer *bytes.Buffer, items []string) error {
	for _, item := range items {
		if _, err := fmt.Fprintln(buffer, item); err != nil {
			return err
		}
	}
	return nil
}

func WriteListToConsole(items []string) error {
	// Формат списка дублируется из-за другого места назначения.
	for _, item := range items {
		if _, err := fmt.Fprintln(os.Stdout, item); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	items := []string{"milk", "bread", "tea"}
	var buffer bytes.Buffer
	if err := WriteListToBuffer(&buffer, items); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Весь список:")
	if err := WriteListToConsole(items); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Первые 11 байт:")
	fmt.Print(buffer.String()[:11]) // Привязано к буферу и длине данных.
}
