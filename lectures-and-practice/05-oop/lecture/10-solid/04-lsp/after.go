package main

import "fmt"

type Flyer interface {
	Fly()
}

type Walker interface {
	Walk()
}

type Sparrow struct{}

func (Sparrow) Fly()  { fmt.Println("Воробей летит") }
func (Sparrow) Walk() { fmt.Println("Воробей идёт") }

type Penguin struct{}

func (Penguin) Walk() { fmt.Println("Пингвин идёт") }

func Launch(bird Flyer) { bird.Fly() }

func WalkTogether(birds ...Walker) {
	for _, bird := range birds {
		bird.Walk()
	}
}
