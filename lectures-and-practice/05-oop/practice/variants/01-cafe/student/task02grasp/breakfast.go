package task02grasp

import (
	"fmt"
	"html"
)

type Component struct {
	Name  string
	Price int
}

type Breakfast struct {
	Name     string
	Coffee   Component
	Sandwich Component
	Fruit    Component
}

func (p Breakfast) HTMLReceipt(quantity int) string {
	total := (p.Coffee.Price + p.Sandwich.Price + p.Fruit.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product Breakfast, quantity int) string {
	total := (product.Coffee.Price + product.Sandwich.Price + product.Fruit.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
