package task03interface

import (
	"testing"

	"example.com/oop-practice/01-cafe/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewCheckout(&clients.Barista{})
	quote, err := service.QuotePreparation("Американо", 3)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 250 || quote.Minutes != 10 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewCheckout(&clients.Barista{})
	if _, err := service.QuotePreparation("Капучино", 3); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
