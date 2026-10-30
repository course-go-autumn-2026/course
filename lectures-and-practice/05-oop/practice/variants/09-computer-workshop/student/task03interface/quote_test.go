package task03interface

import (
	"testing"

	"example.com/oop-practice/09-computer-workshop/student/task03interface/clients"
)

func TestInitialQuote(t *testing.T) {
	service := NewReception(&clients.LocalTechnician{})
	quote, err := service.QuoteService("Ноутбук", 2)
	if err != nil {
		t.Fatal(err)
	}
	if quote.PriceRubles != 1200 || quote.Minutes != 60 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestUnavailableOption(t *testing.T) {
	service := NewReception(&clients.LocalTechnician{})
	if _, err := service.QuoteService("Планшет", 2); err == nil {
		t.Fatal("unavailable option must return an error")
	}
}
