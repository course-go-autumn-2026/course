package task03interface

import "example.com/oop-practice/04-bike-rental/student/task03interface/clients"

type RentalQuote struct {
	PriceRubles int
	Minutes     int
}

type RentalDesk struct {
	client *clients.ParkRental
}

func NewRentalDesk(client *clients.ParkRental) *RentalDesk {
	return &RentalDesk{client: client}
}

func (s *RentalDesk) QuoteRental(bike string, hours int) (RentalQuote, error) {
	price, minutes, err := s.client.Estimate(bike, hours)
	if err != nil {
		return RentalQuote{}, err
	}
	return RentalQuote{PriceRubles: price, Minutes: minutes}, nil
}
