//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: источник данных жёстко связан с конкретными получателями.
// Observer позволит подключать получателей при сборке приложения.

package main

import "fmt"

func ShowOnDisplay(temperature int) {
	fmt.Printf("Табло: %d °C\n", temperature)
}

func WarnAboutHeat(temperature int) {
	if temperature >= 30 {
		fmt.Println("Предупреждение: жарко")
	}
}

type WeatherStation struct{}

func (WeatherStation) SetTemperature(temperature int) {
	ShowOnDisplay(temperature)
	WarnAboutHeat(temperature)
	// Для нового потребителя придётся менять код самой станции.
}

func main() {
	station := WeatherStation{}
	station.SetTemperature(22)
	station.SetTemperature(32)
}
