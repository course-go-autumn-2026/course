package task02grasp

import (
	"fmt"
	"html"
)

type Procedure struct {
	Name  string
	Price int
}

type WashPackage struct {
	Name     string
	Body     Procedure
	Interior Procedure
	Mats     Procedure
}

func (p WashPackage) HTMLReceipt(quantity int) string {
	total := (p.Body.Price + p.Interior.Price + p.Mats.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product WashPackage, quantity int) string {
	total := (product.Body.Price + product.Interior.Price + product.Mats.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
