//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: вложенные ветки дублируют правила цены.
// Решение: ранние возвраты и явный приоритет правил. Дополнительный паттерн не нужен.

package main

import (
	"errors"
	"fmt"
)

// До: правила цены спрятаны во вложенных ветках и частично повторяются.
func TicketPriceBefore(age int, student bool) (int, error) {
	//gocognit
	if age >= 0 {
		if age < 12 {
			return 100, nil
		} else {
			if student {
				if age < 18 {
					return 100, nil
				} else {
					return 150, nil
				}
			} else {
				if age < 18 {
					return 100, nil
				} else {
					return 200, nil
				}
			}
		}
	} else {
		return 0, errors.New("возраст не может быть отрицательным")
	}
}

func main() {
	for _, visitor := range []struct {
		age     int
		student bool
	}{{10, false}, {16, true}, {20, true}, {30, false}, {-1, false}} {
		price, err := TicketPriceBefore(visitor.age, visitor.student)
		fmt.Printf("Возраст %d, студент %t: цена=%d, ошибка=%v\n", visitor.age, visitor.student, price, err)
	}
}
