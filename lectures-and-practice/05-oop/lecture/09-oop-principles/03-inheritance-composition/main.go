// Слайд 9: «Наследование и композиция».
// Запуск из lecture: go run ./09-oop-principles/03-inheritance-composition
// Вопрос: становится ли автомобиль двигателем при встраивании Engine?
package main

import "fmt"

type Engine struct{}

func (Engine) Name() string { return "двигатель" }

func (e Engine) Start() {
	fmt.Println("Запускается:", e.Name())
}

// Композиция с именованным полем: автомобиль содержит двигатель.
type PetrolCar struct {
	engine Engine
}

func (car PetrolCar) Start() {
	fmt.Println("Автомобиль проверяет готовность")
	car.engine.Start() // Явное делегирование части работы.
}

// Встраивание тоже создаёт поле. Метод Start продвигается во внешний тип.
type EmbeddedCar struct {
	Engine
}

// Скрывает продвинутый Name при обращении car.Name().
// Внутри Engine.Start получателем остаётся Engine.
func (EmbeddedCar) Name() string { return "автомобиль" }

func InspectEngine(engine Engine) {
	fmt.Println("Осмотр:", engine.Name())
}

func main() {
	car := PetrolCar{engine: Engine{}}
	car.Start()

	embedded := EmbeddedCar{Engine: Engine{}}
	fmt.Println("Имя внешнего типа:", embedded.Name())
	embedded.Start()        // Вызывает Engine.Start, внутри него - Engine.Name.
	embedded.Engine.Start() // Явный доступ к тому же компоненту.

	InspectEngine(embedded.Engine)
	// InspectEngine(embedded) // Ошибка: EmbeddedCar не является Engine.
	// В Go нет классического наследования структур и виртуального
	// переопределения методов через встраивание.
}
