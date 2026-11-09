//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import "fmt"

// Приложение ожидает температуру в градусах Цельсия.
type TemperatureReader interface {
	Celsius() float64
}

// Допустим, этот датчик предоставлен чужой библиотекой и менять его нельзя.
type FahrenheitSensor struct{}

func (FahrenheitSensor) ReadFahrenheit() float64 {
	return 73.4
}

// Adapter меняет интерфейс и единицы измерения чужого компонента.
type SensorAdapter struct {
	Sensor FahrenheitSensor
}

func (a SensorAdapter) Celsius() float64 {
	return (a.Sensor.ReadFahrenheit() - 32) * 5 / 9
}

func ShowTemperature(reader TemperatureReader) {
	fmt.Printf("В комнате %.1f °C\n", reader.Celsius())
}

func IsHot(reader TemperatureReader) bool {
	return reader.Celsius() >= 30
}

func main() {
	sensor := SensorAdapter{Sensor: FahrenheitSensor{}}
	ShowTemperature(sensor)
	fmt.Println("Жарко:", IsHot(sensor))
}
