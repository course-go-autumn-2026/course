//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import "fmt"

// State: одна и та же команда зависит от текущего состояния турникета.
type State interface {
	Coin() State
	Push() State
}

type Locked struct{}

func (Locked) Coin() State {
	fmt.Println("Монета принята: проход открыт")
	return Unlocked{}
}

func (s Locked) Push() State {
	fmt.Println("Проход закрыт: сначала внесите монету")
	return s
}

type Unlocked struct{}

func (s Unlocked) Coin() State {
	fmt.Println("Проход уже открыт: монета возвращена")
	return s
}

func (Unlocked) Push() State {
	fmt.Println("Посетитель прошёл: турникет закрыт")
	return Locked{}
}

type Turnstile struct {
	state State
}

func NewTurnstile() *Turnstile {
	return &Turnstile{state: Locked{}}
}

func (t *Turnstile) Coin() {
	t.state = t.state.Coin()
}

func (t *Turnstile) Push() {
	t.state = t.state.Push()
}

func main() {
	t := NewTurnstile()
	t.Push()
	t.Coin()
	t.Coin()
	t.Push()
	t.Push()
}
