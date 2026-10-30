// Слайд 21: KISS. Запуск из lecture: go run ./11-kiss-yagni-dry/01-kiss
// Вопрос: можно ли упростить чтение проверок, сохранив поведение?
package main

import "fmt"

func main() {
	cases := []struct {
		name  string
		order Order
	}{
		{"пустой заказ", Order{}},
		{"адрес из пробелов", Order{ItemsCount: 2, Address: "   ", Paid: true}},
		{"не оплачен", Order{ItemsCount: 2, Address: "ул. Пушкина, 10"}},
		{"готов к отправке", Order{ItemsCount: 2, Address: "ул. Пушкина, 10", Paid: true}},
	}
	for _, example := range cases {
		fmt.Printf("%s: до = %v; после = %v\n", example.name,
			BeforeValidateDispatch(example.order), ValidateDispatch(example.order))
	}
}
