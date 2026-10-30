package task02grasp

import (
	"fmt"
	"html"
)

type Component struct {
	Name  string
	Price int
}

type MovieSet struct {
	Name    string
	Ticket  Component
	Popcorn Component
	Glasses Component
}

func (p MovieSet) HTMLReceipt(quantity int) string {
	total := (p.Ticket.Price + p.Popcorn.Price + p.Glasses.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product MovieSet, quantity int) string {
	total := (product.Ticket.Price + product.Popcorn.Price + product.Glasses.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
