// Слайд 8: «Полиморфизм».
// Запуск из lecture: go run ./09-oop-principles/02-polymorphism
// Вопрос: как один и тот же вызов Drive выбирает разные реализации?
package main

import "fmt"

type Car interface {
	Drive(distanceKm int)
}

type PetrolCar struct{}

func (PetrolCar) Drive(distanceKm int) {
	fmt.Printf("Бензиновый двигатель: едем %d км\n", distanceKm)
}

type ElectricCar struct{}

func (ElectricCar) Drive(distanceKm int) {
	fmt.Printf("Электромотор: едем %d км\n", distanceKm)
}

// Оба типа удовлетворяют интерфейсу неявно, без общего базового типа.
var _ Car = PetrolCar{}
var _ Car = ElectricCar{}

func Travel(car Car, distanceKm int) {
	car.Drive(distanceKm)
	// Не нужны switch по типу машины или отдельная функция для каждого типа.
}

func main() {
	cars := []Car{PetrolCar{}, ElectricCar{}}
	for _, car := range cars {
		Travel(car, 15)
	}
}
