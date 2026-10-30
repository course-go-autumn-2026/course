package task03interface

import (
	"testing"

	"example.com/oop-practice/06-pet-grooming/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewReception(&clients.LocalGroomer{})
	quote, err := service.QuoteCare("Кошка", 5)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 700 || quote.Minutes != 60 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewReception(&clients.LocalGroomer{})
	if _, err := service.QuoteCare("Кролик", 5); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
