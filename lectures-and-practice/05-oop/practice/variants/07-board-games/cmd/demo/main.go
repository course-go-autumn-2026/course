package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/07-board-games/student/task01solid"
	grasp "example.com/oop-practice/07-board-games/student/task02grasp"
	estimation "example.com/oop-practice/07-board-games/student/task03interface"
	"example.com/oop-practice/07-board-games/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).Rent(solid.Rental{
		ID: "42", Items: []solid.Item{{Name: "Игра на сутки", Price: 300, Quantity: 2}},
		HandoverKind: "locker", Point: "Парк",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.GameSet{
		Name:    "Игровой набор",
		Game:    grasp.Accessory{Name: "Игра", Price: 300},
		Sleeves: grasp.Accessory{Name: "Протекторы", Price: 100},
		Insert:  grasp.Accessory{Name: "Органайзер", Price: 100},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewRentalDesk(&clients.ClubLibrary{}).QuoteRental("Семейная", 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
