package task03interface

import "example.com/oop-practice/03-cinema/student/task03interface/clients"

type TicketQuote struct {
	PriceRubles int
	Minutes     int
}

type BookingDesk struct {
	client *clients.BoxOffice
}

func NewBookingDesk(client *clients.BoxOffice) *BookingDesk {
	return &BookingDesk{client: client}
}

func (s *BookingDesk) QuoteTickets(show string, seats int) (TicketQuote, error) {
	price, minutes, err := s.client.Estimate(show, seats)
	if err != nil {
		return TicketQuote{}, err
	}
	return TicketQuote{PriceRubles: price, Minutes: minutes}, nil
}
