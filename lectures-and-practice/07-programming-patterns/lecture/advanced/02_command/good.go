//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import "fmt"

type Lamp struct {
	on bool
}

func (l *Lamp) SetOn(on bool) {
	l.on = on
	fmt.Println("Лампа включена:", l.on)
}

type Command interface {
	Execute()
}

// Command хранит получателя и аргументы действия.
// Команду можно создать сейчас, а выполнить позже.
type SwitchCommand struct {
	Lamp *Lamp
	On   bool
}

func (c SwitchCommand) Execute() {
	c.Lamp.SetOn(c.On)
}

func main() {
	lamp := &Lamp{}
	queue := []Command{
		SwitchCommand{Lamp: lamp, On: true},
		SwitchCommand{Lamp: lamp, On: false},
		SwitchCommand{Lamp: lamp, On: true},
	}

	fmt.Println("До выполнения очереди:", lamp.on)
	for _, command := range queue {
		command.Execute()
	}
	// Эта очередь живёт только в памяти. Command сам по себе
	// не обеспечивает сохранность очереди, повторы или отмену действий.
}
