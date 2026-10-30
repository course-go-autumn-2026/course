package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/10-print-shop/student/task01solid"
	grasp "example.com/oop-practice/10-print-shop/student/task02grasp"
	estimation "example.com/oop-practice/10-print-shop/student/task03interface"
	"example.com/oop-practice/10-print-shop/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).PlaceOrder(solid.Order{
		ID: "42", Items: []solid.Item{{Name: "Комплект страниц", Price: 300, Quantity: 2}},
		BindingKind: "spiral", Format: "A3",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.PrintSet{
		Name:    "Учебный комплект",
		Pages:   grasp.ServicePart{Name: "Страницы", Price: 300},
		Cover:   grasp.ServicePart{Name: "Обложка", Price: 200},
		Binding: grasp.ServicePart{Name: "Переплёт", Price: 100},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewCheckout(&clients.OfficePrinter{}).QuotePrint("A4", 25)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
