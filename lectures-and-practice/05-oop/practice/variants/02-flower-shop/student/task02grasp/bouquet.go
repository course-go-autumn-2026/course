package task02grasp

import (
	"fmt"
	"html"
)

type Material struct {
	Name  string
	Price int
}

type Bouquet struct {
	Name     string
	Flowers  Material
	Greenery Material
	Ribbon   Material
}

func (p Bouquet) HTMLReceipt(quantity int) string {
	total := (p.Flowers.Price + p.Greenery.Price + p.Ribbon.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product Bouquet, quantity int) string {
	total := (product.Flowers.Price + product.Greenery.Price + product.Ribbon.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
