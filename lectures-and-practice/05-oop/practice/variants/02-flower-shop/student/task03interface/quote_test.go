package task03interface

import (
	"testing"

	"example.com/oop-practice/02-flower-shop/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewCheckout(&clients.LocalFlorist{})
	quote, err := service.QuoteAssembly("Маленький", 3)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 350 || quote.Minutes != 20 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewCheckout(&clients.LocalFlorist{})
	if _, err := service.QuoteAssembly("Гигантский", 3); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
