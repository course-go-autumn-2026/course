//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: очередь хранит только имена, исполнитель знает все виды действий.
// Command свяжет действие, его аргументы и получателя; исполнителю останется Execute.

package main

import "fmt"

type Lamp struct{ on bool }

func (l *Lamp) SetOn(on bool) {
	l.on = on
	fmt.Println("Лампа включена:", l.on)
}

func main() {
	lamp := &Lamp{}
	queue := []string{"on", "off", "on"}
	fmt.Println("До выполнения очереди:", lamp.on)
	for _, action := range queue {
		// Исполнитель знает команды конкретной лампы и сам ищет получателя.
		switch action {
		case "on":
			lamp.SetOn(true)
		case "off":
			lamp.SetOn(false)
		default:
			fmt.Println("Неизвестное действие:", action)
		}
	}
}
