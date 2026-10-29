package hotel

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"standard", "StandardRoom", 450}, {"suite", "SuiteRoom", 550},
	} {
		name, price, err := legacyBooking(tc.kind, false, false, "direct")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyBooking("unknown", false, false, "direct"); name != "" || price != 0 || !errors.Is(err, ErrUnknownRoom) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		breakfast, parking bool
		name               string
		price              int
	}{
		{false, false, "StandardRoom", 450}, {true, false, "StandardRoom + breakfast", 500},
		{false, true, "StandardRoom + parking", 480}, {true, true, "StandardRoom + breakfast + parking", 530},
	} {
		name, price, err := legacyBooking("standard", tc.breakfast, tc.parking, "direct")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyBooking("unknown", true, true, "direct"); name != "" || price != 0 || !errors.Is(err, ErrUnknownRoom) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyTariff(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"direct", 450}, {"agent", 550}, {"priority", 650}} {
		if _, got, err := legacyBooking("standard", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyBooking("standard", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownTariff) {
		t.Fatal(name, got, err)
	}
}
