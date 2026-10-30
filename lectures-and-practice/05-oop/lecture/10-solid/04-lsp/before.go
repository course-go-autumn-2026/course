package main

import "fmt"

// Договорённость примера: здоровая птица, переданная как Flyer, умеет летать.
// Неподдерживаемая возможность не должна маскироваться подходящей сигнатурой.
type BeforeFlyer interface {
	Fly()
}

type BeforeSparrow struct{}

func (BeforeSparrow) Fly() { fmt.Println("Воробей летит") }

type BeforePenguin struct{}

func (BeforePenguin) Fly() { panic("пингвин не умеет летать") }

// Компилятор проверяет методы, но не смысловое обещание контракта.
var _ BeforeFlyer = BeforePenguin{}

func BeforeLaunch(bird BeforeFlyer) { bird.Fly() }

// recover нужен исключительно для демонстрации нарушения и продолжения
// лекционного примера. Он не исправляет контракт и не является решением LSP.
func ShowBrokenSubstitution() {
	defer func() {
		if reason := recover(); reason != nil {
			fmt.Println("Подстановка пингвина вызвала panic:", reason)
		}
	}()
	BeforeLaunch(BeforePenguin{})
}
