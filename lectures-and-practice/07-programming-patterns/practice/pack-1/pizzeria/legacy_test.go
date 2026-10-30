package pizzeria

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"margherita", "Margherita", 450}, {"pepperoni", "Pepperoni", 550},
	} {
		name, price, err := legacyOrder(tc.kind, false, false, "pickup")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyOrder("unknown", false, false, "pickup"); name != "" || price != 0 || !errors.Is(err, ErrUnknownPizza) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		cheese, box bool
		name        string
		price       int
	}{
		{false, false, "Margherita", 450}, {true, false, "Margherita + cheese", 500},
		{false, true, "Margherita + box", 480}, {true, true, "Margherita + cheese + box", 530},
	} {
		name, price, err := legacyOrder("margherita", tc.cheese, tc.box, "pickup")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyOrder("unknown", true, true, "pickup"); name != "" || price != 0 || !errors.Is(err, ErrUnknownPizza) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyDelivery(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"pickup", 450}, {"courier", 550}, {"express", 650}} {
		if _, got, err := legacyOrder("margherita", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyOrder("margherita", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownDelivery) {
		t.Fatal(name, got, err)
	}
}
