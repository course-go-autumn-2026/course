package task03interface

import "example.com/oop-practice/01-cafe/student/task03interface/clients"

type PreparationQuote struct {
	PriceRubles int
	Minutes     int
}

type Checkout struct {
	client *clients.Barista
}

func NewCheckout(client *clients.Barista) *Checkout {
	return &Checkout{client: client}
}

func (s *Checkout) QuotePreparation(drink string, cups int) (PreparationQuote, error) {
	price, minutes, err := s.client.Estimate(drink, cups)
	if err != nil {
		return PreparationQuote{}, err
	}
	return PreparationQuote{PriceRubles: price, Minutes: minutes}, nil
}
