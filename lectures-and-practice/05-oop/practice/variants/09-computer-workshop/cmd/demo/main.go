package main

import (
	"fmt"
	"log"

	solid "example.com/oop-practice/09-computer-workshop/student/task01solid"
	grasp "example.com/oop-practice/09-computer-workshop/student/task02grasp"
	estimation "example.com/oop-practice/09-computer-workshop/student/task03interface"
	"example.com/oop-practice/09-computer-workshop/student/task03interface/clients"
)

func main() {
	receipt, err := solid.NewService(solid.NewMemoryStore()).Accept(solid.Request{
		ID: "42", Items: []solid.Item{{Name: "Установка памяти", Price: 500, Quantity: 2}},
		DiagnosticsKind: "standard", Device: "Компьютер",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt)

	product := grasp.UpgradeKit{
		Name:         "Обновление компьютера",
		Memory:       grasp.Part{Name: "Память", Price: 2000},
		Storage:      grasp.Part{Name: "Накопитель", Price: 3000},
		Installation: grasp.Part{Name: "Установка", Price: 500},
	}
	fmt.Println((grasp.Receipt{}).Text(product, 2))

	quote, err := estimation.NewReception(&clients.LocalTechnician{}).QuoteService("Ноутбук", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Оценка: %d руб., %d мин.\n", quote.PriceRubles, quote.Minutes)
}
