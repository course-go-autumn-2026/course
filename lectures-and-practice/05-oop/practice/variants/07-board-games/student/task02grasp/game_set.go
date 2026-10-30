package task02grasp

import (
	"fmt"
	"html"
)

type Accessory struct {
	Name  string
	Price int
}

type GameSet struct {
	Name    string
	Game    Accessory
	Sleeves Accessory
	Insert  Accessory
}

func (p GameSet) HTMLReceipt(quantity int) string {
	total := (p.Game.Price + p.Sleeves.Price + p.Insert.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product GameSet, quantity int) string {
	total := (product.Game.Price + product.Sleeves.Price + product.Insert.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
