package task02grasp

import (
	"fmt"
	"html"
)

type ServicePart struct {
	Name  string
	Price int
}

type PrintSet struct {
	Name    string
	Pages   ServicePart
	Cover   ServicePart
	Binding ServicePart
}

func (p PrintSet) HTMLReceipt(quantity int) string {
	total := (p.Pages.Price + p.Cover.Price + p.Binding.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product PrintSet, quantity int) string {
	total := (product.Pages.Price + product.Cover.Price + product.Binding.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
