//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: потребители зависят от чужого датчика и повторяют перевод единиц.
// Adapter даст им общий интерфейс температуры в Цельсиях.

package main

import "fmt"

type FahrenheitSensor struct{}

func (FahrenheitSensor) ReadFahrenheit() float64 { return 73.4 }

func ShowTemperature(sensor FahrenheitSensor) {
	celsius := (sensor.ReadFahrenheit() - 32) * 5 / 9
	fmt.Printf("В комнате %.1f °C\n", celsius)
}

func IsHot(sensor FahrenheitSensor) bool {
	// Второй потребитель снова знает интерфейс датчика и формулу перевода.
	celsius := (sensor.ReadFahrenheit() - 32) * 5 / 9
	return celsius >= 30
}

func main() {
	sensor := FahrenheitSensor{}
	ShowTemperature(sensor)
	fmt.Println("Жарко:", IsHot(sensor))
}
