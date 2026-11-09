//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: вложенные ветки дублируют правила цены.
// Решение: ранние возвраты и явный приоритет правил. Дополнительный паттерн не нужен.

package main

import (
	"errors"
	"fmt"
)

// После: сначала неверный ввод, затем правила в порядке приоритета.
func TicketPriceAfter(age int, student bool) (int, error) {
	if age < 0 {
		return 0, errors.New("возраст не может быть отрицательным")
	}
	if age < 18 {
		return 100, nil
	}
	if student {
		return 150, nil
	}
	return 200, nil
}

func main() {
	for _, visitor := range []struct {
		age     int
		student bool
	}{{10, false}, {16, true}, {20, true}, {30, false}, {-1, false}} {
		price, err := TicketPriceAfter(visitor.age, visitor.student)
		fmt.Printf("Возраст %d, студент %t: цена=%d, ошибка=%v\n", visitor.age, visitor.student, price, err)
	}
}
