package hotel

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
		p     Room
		name  string
		price int
	}{
		{StandardRoom{}, "StandardRoom", 450}, {SuiteRoom{}, "SuiteRoom", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"standard", "suite"} {
		name, price, _ := legacyBooking(kind, false, false, "direct")
		got, err := NewRoom(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "StandardRoom"} {
		p, err := NewRoom(kind)
		if p != nil || !errors.Is(err, ErrUnknownRoom) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"standard", "suite"} {
		for _, breakfast := range []bool{false, true} {
			for _, parking := range []bool{false, true} {
				baseName, basePrice, _ := legacyBooking(kind, false, false, "direct")
				var p Room = &productProbe{name: baseName, price: basePrice}
				if breakfast {
					p = WithBreakfast{Base: p}
				}
				if parking {
					p = WithParking{Base: p}
				}
				name, price, _ := legacyBooking(kind, breakfast, parking, "direct")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, breakfast, parking, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(Room) Room
		suffix string
		extra  int
	}{
		{func(p Room) Room { return WithBreakfast{Base: p} }, " + breakfast", 50},
		{func(p Room) Room { return WithParking{Base: p} }, " + parking", 30},
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
	p := WithBreakfast{Base: WithParking{Base: base}}
	if p.Name() != "Test + parking + breakfast" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithBreakfast{Base: WithBreakfast{Base: base}}
	if double.Name() != "Test + breakfast + breakfast" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if DirectBooking(base) != base || AgentBooking(base) != base+100 || PriorityBooking(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   BookingPrice
	}{
		{"direct", DirectBooking}, {"agent", AgentBooking}, {"priority", PriorityBooking},
	} {
		for _, kind := range []string{"standard", "suite"} {
			for _, breakfast := range []bool{false, true} {
				for _, parking := range []bool{false, true} {
					_, base, _ := legacyBooking(kind, breakfast, parking, "direct")
					_, want, _ := legacyBooking(kind, breakfast, parking, tc.name)
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
