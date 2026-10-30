package task03interface

import (
	"testing"

	"example.com/oop-practice/03-cinema/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewBookingDesk(&clients.BoxOffice{})
	quote, err := service.QuoteTickets("Дневной", 2)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 700 || quote.Minutes != 5 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewBookingDesk(&clients.BoxOffice{})
	if _, err := service.QuoteTickets("Ночной", 2); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
