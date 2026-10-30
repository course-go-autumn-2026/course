package task02grasp

import (
	"fmt"
	"html"
)

type Equipment struct {
	Name  string
	Price int
}

type RideKit struct {
	Name   string
	Bike   Equipment
	Helmet Equipment
	Lock   Equipment
}

func (p RideKit) HTMLReceipt(quantity int) string {
	total := (p.Bike.Price + p.Helmet.Price + p.Lock.Price) * quantity
	return fmt.Sprintf("<p>%s × %d: %d руб.</p>", html.EscapeString(p.Name), quantity, total)
}

type Receipt struct{}

func (Receipt) Text(product RideKit, quantity int) string {
	total := (product.Bike.Price + product.Helmet.Price + product.Lock.Price) * quantity
	return fmt.Sprintf("%s × %d: %d руб.", product.Name, quantity, total)
}
