package task02grasp

import (
	"fmt"
	"html"
)

type ServicePart struct {
	Name  string
	Price int
}

type PhotoPackage struct {
	Name     string
	Shooting ServicePart
	Retouch  ServicePart
	Prints   ServicePart
}

func (p PhotoPackage) HTMLReceipt(quantity int) string {
	total := (p.Shooting.Price + p.Retouch.Price + p.Prints.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product PhotoPackage, quantity int) string {
	total := (product.Shooting.Price + product.Retouch.Price + product.Prints.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
