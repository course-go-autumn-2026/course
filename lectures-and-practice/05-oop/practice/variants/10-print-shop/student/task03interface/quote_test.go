package task03interface

import (
	"testing"

	"example.com/oop-practice/10-print-shop/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewCheckout(&clients.OfficePrinter{})
	quote, err := service.QuotePrint("A4", 25)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 350 || quote.Minutes != 15 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewCheckout(&clients.OfficePrinter{})
	if _, err := service.QuotePrint("A5", 25); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
