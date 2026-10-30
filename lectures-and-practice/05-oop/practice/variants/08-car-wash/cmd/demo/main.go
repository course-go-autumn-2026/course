package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/08-car-wash/student/task01solid"
	grasp "example.com/oop-practice/08-car-wash/student/task02grasp"
	estimation "example.com/oop-practice/08-car-wash/student/task03interface"
	"example.com/oop-practice/08-car-wash/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).PlaceOrder(solid.Order{
		ID: "42", Items: []solid.Item{{Name: "Мойка кузова", Price: 600, Quantity: 2}},
		FinishKind: "wax", Box: "Бокс 2",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.WashPackage{
		Name:     "Комплексная мойка",
		Body:     grasp.Procedure{Name: "Кузов", Price: 300},
		Interior: grasp.Procedure{Name: "Салон", Price: 200},
		Mats:     grasp.Procedure{Name: "Коврики", Price: 100},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewReception(&clients.ManualWash{}).QuoteWash("Легковой", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
