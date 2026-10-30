package warehouse

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"small", "SmallCell", 450}, {"large", "LargeCell", 550},
	} {
		name, price, err := legacyStorage(tc.kind, false, false, "self")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyStorage("unknown", false, false, "self"); name != "" || price != 0 || !errors.Is(err, ErrUnknownStorageCell) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		climateControl, insurance bool
		name                      string
		price                     int
	}{
		{false, false, "SmallCell", 450}, {true, false, "SmallCell + climate control", 500},
		{false, true, "SmallCell + insurance", 480}, {true, true, "SmallCell + climate control + insurance", 530},
	} {
		name, price, err := legacyStorage("small", tc.climateControl, tc.insurance, "self")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyStorage("unknown", true, true, "self"); name != "" || price != 0 || !errors.Is(err, ErrUnknownStorageCell) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyTariff(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"self", 450}, {"standard", 550}, {"express", 650}} {
		if _, got, err := legacyStorage("small", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyStorage("small", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownTariff) {
		t.Fatal(name, got, err)
	}
}
