package fitness

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"daytime", "DaytimePass", 450}, {"full_day", "FullDayPass", 550},
	} {
		name, price, err := legacyMembership(tc.kind, false, false, "next_week")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyMembership("unknown", false, false, "next_week"); name != "" || price != 0 || !errors.Is(err, ErrUnknownMembership) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		locker, towel bool
		name          string
		price         int
	}{
		{false, false, "DaytimePass", 450}, {true, false, "DaytimePass + locker", 500},
		{false, true, "DaytimePass + towel", 480}, {true, true, "DaytimePass + locker + towel", 530},
	} {
		name, price, err := legacyMembership("daytime", tc.locker, tc.towel, "next_week")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyMembership("unknown", true, true, "next_week"); name != "" || price != 0 || !errors.Is(err, ErrUnknownMembership) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyTariff(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"next_week", 450}, {"today", 550}, {"personal", 650}} {
		if _, got, err := legacyMembership("daytime", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyMembership("daytime", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownTariff) {
		t.Fatal(name, got, err)
	}
}
