//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import (
	"fmt"
	"iter"
)

// Требуется Go 1.23 или новее: range по функции появился в Go 1.23.
// Iterator выдаёт подходящие элементы по одному без промежуточного среза.
func EvenNumbers(numbers []int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, number := range numbers {
			fmt.Println("Проверяем:", number)
			if number%2 == 0 {
				if !yield(number) {
					return // После break потребителя нельзя вызывать yield снова.
				}
			}
		}
	}
}

func main() {
	numbers := []int{1, 2, 3, 4, 5, 6}
	fmt.Println("Чётные числа до ранней остановки:")
	for number := range EvenNumbers(numbers) {
		fmt.Println(number)
		if number == 4 {
			break
		}
	}
}
