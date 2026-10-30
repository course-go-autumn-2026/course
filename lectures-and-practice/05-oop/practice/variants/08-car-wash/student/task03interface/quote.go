package task03interface

import "example.com/oop-practice/08-car-wash/student/task03interface/clients"

type WashQuote struct {
	PriceRubles int
	Minutes     int
}

type Reception struct {
	client *clients.ManualWash
}

func NewReception(client *clients.ManualWash) *Reception {
	return &Reception{client: client}
}

func (s *Reception) QuoteWash(carClass string, cars int) (WashQuote, error) {
	price, minutes, err := s.client.Estimate(carClass, cars)
	if err != nil {
		return WashQuote{}, err
	}
	return WashQuote{PriceRubles: price, Minutes: minutes}, nil
}
