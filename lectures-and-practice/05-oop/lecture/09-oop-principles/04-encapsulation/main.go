// Слайд 10: «Инкапсуляция».
// Запуск из lecture: go run ./09-oop-principles/04-encapsulation
// Вопрос: достаточно ли скрыть speed, если SetSpeed принимает любое число?
package main

import (
	"fmt"

	"oop-lecture-examples/09-oop-principles/04-encapsulation/vehicle"
)

func main() {
	car := vehicle.PetrolCar{}
	fmt.Println("Начальная скорость:", car.Speed())

	for _, speed := range []int{80, -10, 200, 180, 0} {
		err := car.SetSpeed(speed)
		fmt.Printf("SetSpeed(%d): %v; скорость: %d\n", speed, err, car.Speed())
	}

	// Раскомментировать для демонстрации ошибки компиляции:
	// car.speed = -10 // main не имеет доступа к полю из пакета vehicle.
	// Скрытие поля и проверка правил в методе вместе защищают состояние.
}
