package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/03-cinema/student/task01solid"
	grasp "example.com/oop-practice/03-cinema/student/task02grasp"
	estimation "example.com/oop-practice/03-cinema/student/task03interface"
	"example.com/oop-practice/03-cinema/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).Book(solid.Booking{
		ID: "42", Items: []solid.Item{{Name: "Билет", Price: 400, Quantity: 2}},
		SeatKind: "comfort", Hall: "Малый зал",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.MovieSet{
		Name:    "Киновечер",
		Ticket:  grasp.Component{Name: "Билет", Price: 300},
		Popcorn: grasp.Component{Name: "Попкорн", Price: 200},
		Glasses: grasp.Component{Name: "3D-очки", Price: 100},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewBookingDesk(&clients.BoxOffice{}).QuoteTickets("Дневной", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
