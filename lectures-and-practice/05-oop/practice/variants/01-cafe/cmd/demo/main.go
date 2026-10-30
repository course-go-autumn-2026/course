package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/01-cafe/student/task01solid"
	grasp "example.com/oop-practice/01-cafe/student/task02grasp"
	estimation "example.com/oop-practice/01-cafe/student/task03interface"
	"example.com/oop-practice/01-cafe/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).PlaceOrder(solid.Order{
		ID: "42", Items: []solid.Item{{Name: "Латте", Price: 250, Quantity: 2}},
		ServingKind: "table", Area: "Зал",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.Breakfast{
		Name:     "Завтрак",
		Coffee:   grasp.Component{Name: "Кофе", Price: 150},
		Sandwich: grasp.Component{Name: "Сэндвич", Price: 250},
		Fruit:    grasp.Component{Name: "Фрукты", Price: 100},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewCheckout(&clients.Barista{}).QuotePreparation("Американо", 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
