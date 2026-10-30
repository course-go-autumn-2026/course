package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/04-bike-rental/student/task01solid"
	grasp "example.com/oop-practice/04-bike-rental/student/task02grasp"
	estimation "example.com/oop-practice/04-bike-rental/student/task03interface"
	"example.com/oop-practice/04-bike-rental/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).Rent(solid.Rental{
		ID: "42", Items: []solid.Item{{Name: "Велосипед на час", Price: 300, Quantity: 2}},
		ReturnKind: "other", Station: "Парк",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.RideKit{
		Name:   "Набор для прогулки",
		Bike:   grasp.Equipment{Name: "Велосипед", Price: 500},
		Helmet: grasp.Equipment{Name: "Шлем", Price: 100},
		Lock:   grasp.Equipment{Name: "Замок", Price: 50},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewRentalDesk(&clients.ParkRental{}).QuoteRental("Городской", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
