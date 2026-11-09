//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Accept(Visitor)
}

type Visitor interface {
	VisitCircle(Circle)
	VisitRectangle(Rectangle)
}

type Circle struct {
	Radius float64
}

func (c Circle) Accept(v Visitor) {
	v.VisitCircle(c)
}

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Accept(v Visitor) {
	v.VisitRectangle(r)
}

// Visitor позволяет добавлять операции без изменения типов фигур.
type Printer struct{}

func (Printer) VisitCircle(c Circle) {
	fmt.Printf("Круг: радиус %.0f\n", c.Radius)
}

func (Printer) VisitRectangle(r Rectangle) {
	fmt.Printf("Прямоугольник: %.0f x %.0f\n", r.Width, r.Height)
}

type AreaCounter struct {
	Total float64
}

func (a *AreaCounter) VisitCircle(c Circle) {
	a.Total += math.Pi * c.Radius * c.Radius
}

func (a *AreaCounter) VisitRectangle(r Rectangle) {
	a.Total += r.Width * r.Height
}

func main() {
	shapes := []Shape{Circle{Radius: 2}, Rectangle{Width: 3, Height: 4}}
	area := &AreaCounter{}
	for _, shape := range shapes {
		shape.Accept(Printer{})
		shape.Accept(area)
	}
	fmt.Printf("Общая площадь: %.2f\n", area.Total)
	// Цена подхода: новый вид фигуры потребует расширить Visitor
	// и дописать соответствующий метод у каждого посетителя.
}
