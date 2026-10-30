package shipping

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"envelope", "Envelope", 450}, {"parcel", "Parcel", 550},
	} {
		name, price, err := legacyShipment(tc.kind, false, false, "drop_off")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyShipment("unknown", false, false, "drop_off"); name != "" || price != 0 || !errors.Is(err, ErrUnknownShipment) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		insurance, fragileHandling bool
		name                       string
		price                      int
	}{
		{false, false, "Envelope", 450}, {true, false, "Envelope + insurance", 500},
		{false, true, "Envelope + fragile handling", 480}, {true, true, "Envelope + insurance + fragile handling", 530},
	} {
		name, price, err := legacyShipment("envelope", tc.insurance, tc.fragileHandling, "drop_off")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyShipment("unknown", true, true, "drop_off"); name != "" || price != 0 || !errors.Is(err, ErrUnknownShipment) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyTariff(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"drop_off", 450}, {"courier_pickup", 550}, {"express_pickup", 650}} {
		if _, got, err := legacyShipment("envelope", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyShipment("envelope", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownTariff) {
		t.Fatal(name, got, err)
	}
}
