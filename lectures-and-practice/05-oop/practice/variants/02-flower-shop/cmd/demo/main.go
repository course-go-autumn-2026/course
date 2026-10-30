package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/02-flower-shop/student/task01solid"
	grasp "example.com/oop-practice/02-flower-shop/student/task02grasp"
	estimation "example.com/oop-practice/02-flower-shop/student/task03interface"
	"example.com/oop-practice/02-flower-shop/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).PlaceOrder(solid.Order{
		ID: "42", Items: []solid.Item{{Name: "Букет", Price: 550, Quantity: 2}},
		PackagingKind: "box", Size: "Маленький",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.Bouquet{
		Name:     "Весенний букет",
		Flowers:  grasp.Material{Name: "Цветы", Price: 400},
		Greenery: grasp.Material{Name: "Зелень", Price: 100},
		Ribbon:   grasp.Material{Name: "Лента", Price: 50},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewCheckout(&clients.LocalFlorist{}).QuoteAssembly("Маленький", 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
