package task02grasp

import (
	"fmt"
	"html"
)

type Part struct {
	Name  string
	Price int
}

type UpgradeKit struct {
	Name         string
	Memory       Part
	Storage      Part
	Installation Part
}

func (p UpgradeKit) HTMLReceipt(quantity int) string {
	total := (p.Memory.Price + p.Storage.Price + p.Installation.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product UpgradeKit, quantity int) string {
	total := (product.Memory.Price + product.Storage.Price + product.Installation.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
