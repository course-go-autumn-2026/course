package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/05-photo-studio/student/task01solid"
	grasp "example.com/oop-practice/05-photo-studio/student/task02grasp"
	estimation "example.com/oop-practice/05-photo-studio/student/task03interface"
	"example.com/oop-practice/05-photo-studio/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).Book(solid.Booking{
		ID: "42", Items: []solid.Item{{Name: "Час съёмки", Price: 1000, Quantity: 2}},
		BackdropKind: "color", Room: "Студия B",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.PhotoPackage{
		Name:     "Семейная съёмка",
		Shooting: grasp.ServicePart{Name: "Съёмка", Price: 1000},
		Retouch:  grasp.ServicePart{Name: "Ретушь", Price: 400},
		Prints:   grasp.ServicePart{Name: "Печать фото", Price: 100},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewBookingDesk(&clients.LocalStudio{}).QuoteSession("Студия A", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
