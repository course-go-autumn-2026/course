// Слайд 20: DIP. Запуск из lecture: go run ./10-solid/05-dip
// Вопрос: всегда ли передача зависимости через конструктор означает DIP?
package main

import (
	"errors"
	"fmt"
)

func main() {
	fmt.Println("До: конкретная печь создаётся внутри")
	if err := (CoupledPizzeria{}).BakePizza(); err != nil {
		panic(err)
	}

	fmt.Println("Только DI: конкретная печь передаётся извне")
	concrete := InjectedPizzeria{oven: &ElectricOven{}}
	if err := concrete.BakePizza(); err != nil {
		panic(err)
	}

	fmt.Println("DIP и DI: пиццерия зависит от контракта")
	for _, oven := range []Oven{ElectricOven{}, WoodOven{}} {
		if err := NewPizzeria(oven).BakePizza(); err != nil {
			panic(err)
		}
	}

	recording := &RecordingOven{}
	if err := NewPizzeria(recording).BakePizza(); err != nil {
		panic(err)
	}
	fmt.Println("Вызовов подставной печи:", recording.calls)

	failing := &RecordingOven{err: errors.New("печь недоступна")}
	fmt.Println("Ошибка возвращена вызывающему коду:", NewPizzeria(failing).BakePizza())
	fmt.Println("Без зависимости:", NewPizzeria(nil).BakePizza())
}
