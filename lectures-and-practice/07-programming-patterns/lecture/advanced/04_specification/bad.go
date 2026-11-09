//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: список и счётчик независимо реализуют одно условие и дают разные ответы.
// Specification позволит передать обоим одну составную проверку.

package main

import "fmt"

type Product struct {
	Name  string
	Price int
	Stock int
}

func PrintCheapProducts(products []Product) {
	for _, product := range products {
		if product.Price < 500 && product.Stock > 0 {
			fmt.Println(product.Name)
		}
	}
}

func CountCheapProducts(products []Product) int {
	count := 0
	for _, product := range products {
		// Копия правила разошлась с исходником: забыли остаток.
		if product.Price < 500 {
			count++
		}
	}
	return count
}

func main() {
	products := []Product{
		{Name: "Карандаш", Price: 50, Stock: 10},
		{Name: "Блокнот", Price: 250, Stock: 0},
		{Name: "Рюкзак", Price: 3000, Stock: 5},
	}
	fmt.Println("В наличии дешевле 500 рублей:")
	PrintCheapProducts(products)
	fmt.Println("Количество:", CountCheapProducts(products))
}
