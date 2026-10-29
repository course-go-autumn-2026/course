package cinema

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"standard", "StandardTicket", 450}, {"premium", "PremiumTicket", 550},
	} {
		name, price, err := legacyTicket(tc.kind, false, false, "box_office")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyTicket("unknown", false, false, "box_office"); name != "" || price != 0 || !errors.Is(err, ErrUnknownTicket) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		popcorn, drink bool
		name           string
		price          int
	}{
		{false, false, "StandardTicket", 450}, {true, false, "StandardTicket + popcorn", 500},
		{false, true, "StandardTicket + drink", 480}, {true, true, "StandardTicket + popcorn + drink", 530},
	} {
		name, price, err := legacyTicket("standard", tc.popcorn, tc.drink, "box_office")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyTicket("unknown", true, true, "box_office"); name != "" || price != 0 || !errors.Is(err, ErrUnknownTicket) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyTariff(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"box_office", 450}, {"online", 550}, {"concierge", 650}} {
		if _, got, err := legacyTicket("standard", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyTicket("standard", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownTariff) {
		t.Fatal(name, got, err)
	}
}
