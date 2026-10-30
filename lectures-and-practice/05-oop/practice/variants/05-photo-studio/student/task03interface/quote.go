package task03interface

import "example.com/oop-practice/05-photo-studio/student/task03interface/clients"

type SessionQuote struct {
	PriceRubles int
	Minutes     int
}

type BookingDesk struct {
	client *clients.LocalStudio
}

func NewBookingDesk(client *clients.LocalStudio) *BookingDesk {
	return &BookingDesk{client: client}
}

func (s *BookingDesk) QuoteSession(room string, hours int) (SessionQuote, error) {
	price, minutes, err := s.client.Estimate(room, hours)
	if err != nil {
		return SessionQuote{}, err
	}
	return SessionQuote{PriceRubles: price, Minutes: minutes}, nil
}
