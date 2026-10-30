package task02grasp

import (
	"fmt"
	"html"
)

type Procedure struct {
	Name  string
	Price int
}

type CareSet struct {
	Name    string
	Washing Procedure
	Haircut Procedure
	Drying  Procedure
}

func (p CareSet) HTMLReceipt(quantity int) string {
	total := (p.Washing.Price + p.Haircut.Price + p.Drying.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product CareSet, quantity int) string {
	total := (product.Washing.Price + product.Haircut.Price + product.Drying.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
