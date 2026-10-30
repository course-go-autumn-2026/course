package task03interface

import "example.com/oop-practice/02-flower-shop/student/task03interface/clients"

type AssemblyQuote struct {
	PriceRubles int
	Minutes     int
}

type Checkout struct {
	client *clients.LocalFlorist
}

func NewCheckout(client *clients.LocalFlorist) *Checkout {
	return &Checkout{client: client}
}

func (s *Checkout) QuoteAssembly(size string, count int) (AssemblyQuote, error) {
	price, minutes, err := s.client.Estimate(size, count)
	if err != nil {
		return AssemblyQuote{}, err
	}
	return AssemblyQuote{PriceRubles: price, Minutes: minutes}, nil
}
