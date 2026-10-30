package main

import "fmt"

type ElectricOven struct{}

func (ElectricOven) Bake() error {
	fmt.Println("Электропечь: пицца выпекается")
	return nil
}

// Логика пиццерии сама выбирает и создаёт конкретную зависимость.
type CoupledPizzeria struct{}

func (CoupledPizzeria) BakePizza() error {
	oven := ElectricOven{}
	return oven.Bake()
}

// Dependency Injection уже есть: печь передаётся извне.
// Но зависимость от конкретного типа ElectricOven сохраняется.
type InjectedPizzeria struct {
	oven *ElectricOven
}

func (p InjectedPizzeria) BakePizza() error { return p.oven.Bake() }
