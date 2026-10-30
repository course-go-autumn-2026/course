package task03interface

import (
	"testing"

	"example.com/oop-practice/04-bike-rental/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewRentalDesk(&clients.ParkRental{})
	quote, err := service.QuoteRental("Городской", 2)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 400 || quote.Minutes != 15 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewRentalDesk(&clients.ParkRental{})
	if _, err := service.QuoteRental("Тандем", 2); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
