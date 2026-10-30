package task03interface

import (
	"testing"

	"example.com/oop-practice/05-photo-studio/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewBookingDesk(&clients.LocalStudio{})
	quote, err := service.QuoteSession("Студия A", 2)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 1800 || quote.Minutes != 20 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewBookingDesk(&clients.LocalStudio{})
	if _, err := service.QuoteSession("Студия C", 2); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
