package task03interface

import "example.com/oop-practice/10-print-shop/student/task03interface/clients"

type PrintQuote struct {
	PriceRubles int
	Minutes     int
}

type Checkout struct {
	client *clients.OfficePrinter
}

func NewCheckout(client *clients.OfficePrinter) *Checkout {
	return &Checkout{client: client}
}

func (s *Checkout) QuotePrint(format string, pages int) (PrintQuote, error) {
	price, minutes, err := s.client.Estimate(format, pages)
	if err != nil {
		return PrintQuote{}, err
	}
	return PrintQuote{PriceRubles: price, Minutes: minutes}, nil
}
