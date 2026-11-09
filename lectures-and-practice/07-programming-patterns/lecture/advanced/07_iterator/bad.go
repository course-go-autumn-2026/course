//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: потребителю нужны первые два результата, но фильтр заранее обходит весь срез.
// Iterator позволит остановить вычисления вместе с потребителем и не создаст новый срез.
// Для маленького полностью нужного результата возврат среза сам по себе нормален.

package main

import "fmt"

func EvenNumbers(numbers []int) []int {
	result := []int{}
	for _, number := range numbers {
		fmt.Println("Проверяем:", number)
		if number%2 == 0 {
			result = append(result, number)
		}
	}
	return result
}

func main() {
	numbers := []int{1, 2, 3, 4, 5, 6}
	fmt.Println("Чётные числа до ранней остановки:")
	for _, number := range EvenNumbers(numbers) {
		fmt.Println(number)
		if number == 4 {
			break
		}
	}
}
