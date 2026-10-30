package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/06-pet-grooming/student/task01solid"
	grasp "example.com/oop-practice/06-pet-grooming/student/task02grasp"
	estimation "example.com/oop-practice/06-pet-grooming/student/task03interface"
	"example.com/oop-practice/06-pet-grooming/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).Book(solid.Appointment{
		ID: "42", Items: []solid.Item{{Name: "Стрижка", Price: 700, Quantity: 2}},
		CareKind: "coat", Size: "Маленький",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.CareSet{
		Name:    "Полный уход",
		Washing: grasp.Procedure{Name: "Мытьё", Price: 200},
		Haircut: grasp.Procedure{Name: "Стрижка", Price: 400},
		Drying:  grasp.Procedure{Name: "Сушка", Price: 100},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewReception(&clients.LocalGroomer{}).QuoteCare("Кошка", 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
