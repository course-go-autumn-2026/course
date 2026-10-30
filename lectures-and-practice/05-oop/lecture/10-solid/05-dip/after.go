package main

import (
	"errors"
	"fmt"
)

// Контракт принадлежит логике-потребителю: ей нужна возможность выпекать.
// В большом проекте его можно разместить в пакете бизнес-логики,
// отдельно от пакетов конкретных печей.
type Oven interface {
	Bake() error
}

type Pizzeria struct {
	oven Oven
}

func NewPizzeria(oven Oven) Pizzeria { return Pizzeria{oven: oven} }

func (p Pizzeria) BakePizza() error {
	if p.oven == nil {
		return errors.New("печь не настроена")
	}
	return p.oven.Bake()
}

type WoodOven struct{}

func (WoodOven) Bake() error {
	fmt.Println("Дровяная печь: пицца выпекается")
	return nil
}

// Подстановка для проверки сценария без настоящей печи и внешних вызовов.
type RecordingOven struct {
	calls int
	err   error
}

func (o *RecordingOven) Bake() error {
	o.calls++
	return o.err
}
