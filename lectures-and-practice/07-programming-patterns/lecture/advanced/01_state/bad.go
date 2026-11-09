//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: обработчики событий несогласованно проверяют состояние.
// State соберёт поведение каждого состояния в одном типе.
// Для двух состояний достаточно и корректного switch; паттерн полезен при росте правил.

package main

import "fmt"

type Turnstile struct{ locked bool }

func (t *Turnstile) Coin() {
	if t.locked {
		fmt.Println("Монета принята: проход открыт")
		t.locked = false
	} else {
		fmt.Println("Проход уже открыт: монета возвращена")
	}
}

func (t *Turnstile) Push() {
	// Забыли проверить состояние: посетитель проходит даже без монеты.
	fmt.Println("Посетитель прошёл: турникет закрыт")
	t.locked = true
}

func main() {
	t := &Turnstile{locked: true}
	t.Push()
	t.Coin()
	t.Coin()
	t.Push()
	t.Push()
}
