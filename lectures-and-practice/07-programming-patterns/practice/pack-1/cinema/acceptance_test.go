package cinema

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
		p     Ticket
		name  string
		price int
	}{
		{StandardTicket{}, "StandardTicket", 450}, {PremiumTicket{}, "PremiumTicket", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"standard", "premium"} {
		name, price, _ := legacyTicket(kind, false, false, "box_office")
		got, err := NewTicket(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "StandardTicket"} {
		p, err := NewTicket(kind)
		if p != nil || !errors.Is(err, ErrUnknownTicket) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"standard", "premium"} {
		for _, popcorn := range []bool{false, true} {
			for _, drink := range []bool{false, true} {
				baseName, basePrice, _ := legacyTicket(kind, false, false, "box_office")
				var p Ticket = &productProbe{name: baseName, price: basePrice}
				if popcorn {
					p = WithPopcorn{Base: p}
				}
				if drink {
					p = WithDrink{Base: p}
				}
				name, price, _ := legacyTicket(kind, popcorn, drink, "box_office")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, popcorn, drink, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(Ticket) Ticket
		suffix string
		extra  int
	}{
		{func(p Ticket) Ticket { return WithPopcorn{Base: p} }, " + popcorn", 50},
		{func(p Ticket) Ticket { return WithDrink{Base: p} }, " + drink", 30},
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
	p := WithPopcorn{Base: WithDrink{Base: base}}
	if p.Name() != "Test + drink + popcorn" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithPopcorn{Base: WithPopcorn{Base: base}}
	if double.Name() != "Test + popcorn + popcorn" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if BoxOffice(base) != base || OnlineBooking(base) != base+100 || ConciergeBooking(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   TicketPrice
	}{
		{"box_office", BoxOffice}, {"online", OnlineBooking}, {"concierge", ConciergeBooking},
	} {
		for _, kind := range []string{"standard", "premium"} {
			for _, popcorn := range []bool{false, true} {
				for _, drink := range []bool{false, true} {
					_, base, _ := legacyTicket(kind, popcorn, drink, "box_office")
					_, want, _ := legacyTicket(kind, popcorn, drink, tc.name)
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
