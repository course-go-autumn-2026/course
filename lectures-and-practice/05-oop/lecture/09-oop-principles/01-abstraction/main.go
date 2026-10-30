// Слайд 7: «Абстракция».
// Запуск из lecture: go run ./09-oop-principles/01-abstraction
// Вопрос: какие возможности автомобиля нужны функции Travel?
package main

import "fmt"

// Для задачи поездки важна только способность ехать.
// В учебном примере расстояние всегда неотрицательное.
type Car interface {
	Drive(distanceKm int)
}

type PetrolCar struct{}

func (PetrolCar) Drive(distanceKm int) {
	fmt.Printf("Бензиновый двигатель: едем %d км\n", distanceKm)
}

// Конкретный тип умеет больше, чем требуется для поездки.
func (PetrolCar) OpenFuelLid() {
	fmt.Println("Лючок бензобака открыт")
}

func Travel(car Car, distanceKm int) {
	fmt.Println("Начинаем поездку")
	car.Drive(distanceKm)
	// car.OpenFuelLid() // Ошибка: контракт Car не содержит этого метода.
	// Travel не нужно знать устройство двигателя или бензобака.
}

func main() {
	car := PetrolCar{}
	car.OpenFuelLid() // Конкретному значению этот метод доступен.
	Travel(car, 15)
}
