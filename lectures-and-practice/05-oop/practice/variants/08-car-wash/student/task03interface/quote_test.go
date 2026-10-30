package task03interface

import (
	"testing"

	"example.com/oop-practice/08-car-wash/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewReception(&clients.ManualWash{})
	quote, err := service.QuoteWash("Легковой", 2)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 900 || quote.Minutes != 40 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewReception(&clients.ManualWash{})
	if _, err := service.QuoteWash("Грузовой", 2); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
