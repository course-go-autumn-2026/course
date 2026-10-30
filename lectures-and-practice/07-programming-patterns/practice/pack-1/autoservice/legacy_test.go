package autoservice

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"inspection", "Inspection", 450}, {"alignment", "WheelAlignment", 550},
	} {
		name, price, err := legacyService(tc.kind, false, false, "scheduled")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyService("unknown", false, false, "scheduled"); name != "" || price != 0 || !errors.Is(err, ErrUnknownService) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		carWash, prioritySlot bool
		name                  string
		price                 int
	}{
		{false, false, "Inspection", 450}, {true, false, "Inspection + car wash", 500},
		{false, true, "Inspection + priority slot", 480}, {true, true, "Inspection + car wash + priority slot", 530},
	} {
		name, price, err := legacyService("inspection", tc.carWash, tc.prioritySlot, "scheduled")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyService("unknown", true, true, "scheduled"); name != "" || price != 0 || !errors.Is(err, ErrUnknownService) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyTariff(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"scheduled", 450}, {"same_day", 550}, {"night_shift", 650}} {
		if _, got, err := legacyService("inspection", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyService("inspection", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownTariff) {
		t.Fatal(name, got, err)
	}
}
