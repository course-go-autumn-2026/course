//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import "fmt"

type Listener func(temperature int)

type WeatherStation struct {
	listeners []Listener
}

func (s *WeatherStation) Subscribe(listener Listener) {
	s.listeners = append(s.listeners, listener)
}

func (s *WeatherStation) SetTemperature(temperature int) {
	// Observer уведомляет подписчиков, не зная их конкретных типов.
	// Вызовы синхронные: медленный подписчик задержит остальных.
	for _, listener := range s.listeners {
		listener(temperature)
	}
}

func main() {
	station := WeatherStation{}
	// Для простоты подписываемся до начала публикации, в одной горутине.
	station.Subscribe(func(temperature int) {
		fmt.Printf("Табло: %d °C\n", temperature)
	})
	station.Subscribe(func(temperature int) {
		if temperature >= 30 {
			fmt.Println("Предупреждение: жарко")
		}
	})

	station.SetTemperature(22)
	station.SetTemperature(32)
}
