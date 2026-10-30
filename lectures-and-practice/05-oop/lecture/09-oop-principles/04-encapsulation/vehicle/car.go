// Пакет vehicle отделяет внутреннее состояние автомобиля от вызывающего кода.
package vehicle

import "fmt"

type PetrolCar struct {
	speed int // Поле доступно внутри vehicle, но скрыто от других пакетов.
}

// SetSpeed сохраняет инвариант: скорость находится в диапазоне [0, 180].
// Нулевое значение PetrolCar корректно, конструктор не требуется.
func (car *PetrolCar) SetSpeed(speed int) error {
	if speed < 0 || speed > 180 {
		return fmt.Errorf("недопустимая скорость: %d", speed)
	}
	car.speed = speed
	return nil
}

// Возвращаем число, поэтому вызывающий код не получает доступ к полю.
func (car PetrolCar) Speed() int {
	return car.speed
}
