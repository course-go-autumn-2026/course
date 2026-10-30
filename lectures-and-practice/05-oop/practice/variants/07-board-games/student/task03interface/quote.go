package task03interface

import "example.com/oop-practice/07-board-games/student/task03interface/clients"

type RentalQuote struct {
	PriceRubles int
	Minutes     int
}

type RentalDesk struct {
	client *clients.ClubLibrary
}

func NewRentalDesk(client *clients.ClubLibrary) *RentalDesk {
	return &RentalDesk{client: client}
}

func (s *RentalDesk) QuoteRental(category string, days int) (RentalQuote, error) {
	price, minutes, err := s.client.Estimate(category, days)
	if err != nil {
		return RentalQuote{}, err
	}
	return RentalQuote{PriceRubles: price, Minutes: minutes}, nil
}
