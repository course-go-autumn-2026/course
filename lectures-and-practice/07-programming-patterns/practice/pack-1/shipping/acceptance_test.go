package shipping

import (
	"errors"
	"testing"
)

type productProbe struct {
	name                 string
	price, names, prices int
}

func (p *productProbe) Name() string { p.names++; return p.name }
func (p *productProbe) Price() int   { p.prices++; return p.price }

func TestAcceptanceFactoryProducts(t *testing.T) {
	for _, tc := range []struct {
		p     Shipment
		name  string
		price int
	}{
		{Envelope{}, "Envelope", 450}, {Parcel{}, "Parcel", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"envelope", "parcel"} {
		name, price, _ := legacyShipment(kind, false, false, "drop_off")
		got, err := NewShipment(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "Envelope"} {
		p, err := NewShipment(kind)
		if p != nil || !errors.Is(err, ErrUnknownShipment) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"envelope", "parcel"} {
		for _, insurance := range []bool{false, true} {
			for _, fragileHandling := range []bool{false, true} {
				baseName, basePrice, _ := legacyShipment(kind, false, false, "drop_off")
				var p Shipment = &productProbe{name: baseName, price: basePrice}
				if insurance {
					p = WithInsurance{Base: p}
				}
				if fragileHandling {
					p = WithFragileHandling{Base: p}
				}
				name, price, _ := legacyShipment(kind, insurance, fragileHandling, "drop_off")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, insurance, fragileHandling, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(Shipment) Shipment
		suffix string
		extra  int
	}{
		{func(p Shipment) Shipment { return WithInsurance{Base: p} }, " + insurance", 50},
		{func(p Shipment) Shipment { return WithFragileHandling{Base: p} }, " + fragile handling", 30},
	} {
		base := &productProbe{name: "Special", price: 417}
		p := tc.wrap(base)
		if got := p.Name(); got != "Special"+tc.suffix || base.names != 1 || base.prices != 0 {
			t.Fatal(got, base)
		}
		if got := p.Price(); got != 417+tc.extra || base.prices != 1 || base.names != 1 {
			t.Fatal(got, base)
		}
		base.name = "Seasonal"
		base.price = 600
		if p.Name() != "Seasonal"+tc.suffix || p.Price() != 600+tc.extra {
			t.Fatal("base changes were not delegated")
		}
		if base.names != 2 || base.prices != 2 {
			t.Fatal("base called more than once", base)
		}
	}
}
func TestAcceptanceDecoratorComposition(t *testing.T) {
	base := &productProbe{name: "Test", price: 400}
	p := WithInsurance{Base: WithFragileHandling{Base: base}}
	if p.Name() != "Test + fragile handling + insurance" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithInsurance{Base: WithInsurance{Base: base}}
	if double.Name() != "Test + insurance + insurance" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if DropOff(base) != base || CourierPickup(base) != base+100 || ExpressPickup(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   ShippingPrice
	}{
		{"drop_off", DropOff}, {"courier_pickup", CourierPickup}, {"express_pickup", ExpressPickup},
	} {
		for _, kind := range []string{"envelope", "parcel"} {
			for _, insurance := range []bool{false, true} {
				for _, fragileHandling := range []bool{false, true} {
					_, base, _ := legacyShipment(kind, insurance, fragileHandling, "drop_off")
					_, want, _ := legacyShipment(kind, insurance, fragileHandling, tc.name)
					if got := Checkout(base, tc.fn); got != want {
						t.Fatal(tc.name, base, got, want)
					}
				}
			}
		}
	}
}
func TestAcceptanceStrategyExtension(t *testing.T) {
	calls := 0
	custom := func(base int) int {
		calls++
		if base != 780 {
			t.Fatal(base)
		}
		return 999
	}
	if got := Checkout(780, custom); got != 999 || calls != 1 {
		t.Fatal(got, calls)
	}
}
