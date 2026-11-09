//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: операции принимают any, обработка всех вариантов ничем не проверяется.
// Visitor задаст полный набор методов для каждого посетителя.
// Набор фигур предполагается стабильным; добавление нового вида затронет всех посетителей.

package main

import (
	"fmt"
	"math"
)

type Circle struct{ Radius float64 }
type Rectangle struct{ Width, Height float64 }

func PrintShape(shape any) {
	switch s := shape.(type) {
	case Circle:
		fmt.Printf("Круг: радиус %.0f\n", s.Radius)
	case Rectangle:
		fmt.Printf("Прямоугольник: %.0f x %.0f\n", s.Width, s.Height)
	}
}

func Area(shape any) float64 {
	switch s := shape.(type) {
	case Circle:
		return math.Pi * s.Radius * s.Radius
	case Rectangle:
		return s.Width * s.Height
	}
	return 0 // Неизвестный вид фигуры тихо потеряется при расчёте.
}

func main() {
	shapes := []any{Circle{Radius: 2}, Rectangle{Width: 3, Height: 4}}
	total := 0.0
	for _, shape := range shapes {
		PrintShape(shape)
		total += Area(shape)
	}
	fmt.Printf("Общая площадь: %.2f\n", total)
}
