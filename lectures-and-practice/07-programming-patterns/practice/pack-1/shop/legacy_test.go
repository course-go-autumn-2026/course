package shop

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"keyboard", "Keyboard", 450}, {"mouse", "Mouse", 550},
	} {
		name, price, err := legacyPurchase(tc.kind, false, false, "store_pickup")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyPurchase("unknown", false, false, "store_pickup"); name != "" || price != 0 || !errors.Is(err, ErrUnknownProduct) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		warranty, giftWrap bool
		name               string
		price              int
	}{
		{false, false, "Keyboard", 450}, {true, false, "Keyboard + warranty", 500},
		{false, true, "Keyboard + gift wrap", 480}, {true, true, "Keyboard + warranty + gift wrap", 530},
	} {
		name, price, err := legacyPurchase("keyboard", tc.warranty, tc.giftWrap, "store_pickup")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyPurchase("unknown", true, true, "store_pickup"); name != "" || price != 0 || !errors.Is(err, ErrUnknownProduct) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyTariff(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"store_pickup", 450}, {"courier", 550}, {"express", 650}} {
		if _, got, err := legacyPurchase("keyboard", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyPurchase("keyboard", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownTariff) {
		t.Fatal(name, got, err)
	}
}
