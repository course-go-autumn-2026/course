//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import "fmt"

type Product struct {
	Name  string
	Price int // Рубли, целые суммы для учебного примера.
	Stock int
}

// Specification описывает условие, не изменяя товар.
type Spec func(Product) bool

func (s Spec) And(other Spec) Spec {
	return func(product Product) bool {
		return s(product) && other(product)
	}
}

func InStock(product Product) bool {
	return product.Stock > 0
}

func CheaperThan(limit int) Spec {
	return func(product Product) bool {
		return product.Price < limit
	}
}

func PrintProducts(products []Product, matches Spec) {
	for _, product := range products {
		if matches(product) {
			fmt.Println(product.Name)
		}
	}
}

func CountProducts(products []Product, matches Spec) int {
	count := 0
	for _, product := range products {
		if matches(product) {
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
	availableAndCheap := Spec(InStock).And(CheaperThan(500))
	fmt.Println("В наличии дешевле 500 рублей:")
	PrintProducts(products, availableAndCheap)
	fmt.Println("Количество:", CountProducts(products, availableAndCheap))
}
