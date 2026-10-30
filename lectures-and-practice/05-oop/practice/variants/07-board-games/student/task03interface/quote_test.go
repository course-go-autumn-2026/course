package task03interface

import (
	"testing"

	"example.com/oop-practice/07-board-games/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewRentalDesk(&clients.ClubLibrary{})
	quote, err := service.QuoteRental("Семейная", 3)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 400 || quote.Minutes != 10 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewRentalDesk(&clients.ClubLibrary{})
	if _, err := service.QuoteRental("Коллекционная", 3); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
