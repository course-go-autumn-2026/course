//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: универсальные интерфейсы усложняют единственную операцию.
// Решение: обычная функция. Здесь применение ещё одного паттерна было бы лишним.

package main

import "fmt"

// Задача: сложить два числа. Другие операции и источники данных не нужны.
// До: заранее построили универсальный механизм из нескольких абстракций.
type NumberProvider interface {
	Number() int
}

type Constant int

func (c Constant) Number() int {
	return int(c)
}

type Operation interface {
	Apply(a, b int) int
}

type Addition struct{}

func (Addition) Apply(a, b int) int {
	return a + b
}

type Calculation struct {
	Left, Right NumberProvider
	Operation   Operation
}

func (c Calculation) Run() int {
	return c.Operation.Apply(c.Left.Number(), c.Right.Number())
}

func main() {
	calculation := Calculation{Left: Constant(2), Right: Constant(3), Operation: Addition{}}
	fmt.Println(calculation.Run())
}
