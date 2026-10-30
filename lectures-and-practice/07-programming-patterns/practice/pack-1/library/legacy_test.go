package library

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"standard", "StandardLoan", 450}, {"extended", "ExtendedLoan", 550},
	} {
		name, price, err := legacyLoan(tc.kind, false, false, "desk_pickup")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyLoan("unknown", false, false, "desk_pickup"); name != "" || price != 0 || !errors.Is(err, ErrUnknownLoanPlan) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		scanning, protectiveCover bool
		name                      string
		price                     int
	}{
		{false, false, "StandardLoan", 450}, {true, false, "StandardLoan + scanning", 500},
		{false, true, "StandardLoan + protective cover", 480}, {true, true, "StandardLoan + scanning + protective cover", 530},
	} {
		name, price, err := legacyLoan("standard", tc.scanning, tc.protectiveCover, "desk_pickup")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyLoan("unknown", true, true, "desk_pickup"); name != "" || price != 0 || !errors.Is(err, ErrUnknownLoanPlan) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyTariff(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"desk_pickup", 450}, {"home_delivery", 550}, {"priority_delivery", 650}} {
		if _, got, err := legacyLoan("standard", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyLoan("standard", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownTariff) {
		t.Fatal(name, got, err)
	}
}
