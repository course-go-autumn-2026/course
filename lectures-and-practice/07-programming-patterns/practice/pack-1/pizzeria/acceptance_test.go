package pizzeria

import (
	"errors"
	"testing"
)

type pizzaProbe struct {
	name                 string
	price, names, prices int
}

func (p *pizzaProbe) Name() string { p.names++; return p.name }
func (p *pizzaProbe) Price() int   { p.prices++; return p.price }

func TestAcceptanceFactoryProducts(t *testing.T) {
	for _, tc := range []struct {
		p     Pizza
		name  string
		price int
	}{
		{Margherita{}, "Margherita", 450}, {Pepperoni{}, "Pepperoni", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"margherita", "pepperoni"} {
		name, price, _ := legacyOrder(kind, false, false, "pickup")
		got, err := NewPizza(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "Margherita"} {
		p, err := NewPizza(kind)
		if p != nil || !errors.Is(err, ErrUnknownPizza) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"margherita", "pepperoni"} {
		for _, cheese := range []bool{false, true} {
			for _, box := range []bool{false, true} {
				baseName, basePrice, _ := legacyOrder(kind, false, false, "pickup")
				var p Pizza = &pizzaProbe{name: baseName, price: basePrice}
				if cheese {
					p = WithCheese{Base: p}
				}
				if box {
					p = WithBox{Base: p}
				}
				name, price, _ := legacyOrder(kind, cheese, box, "pickup")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, cheese, box, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(Pizza) Pizza
		suffix string
		extra  int
	}{
		{func(p Pizza) Pizza { return WithCheese{Base: p} }, " + cheese", 50},
		{func(p Pizza) Pizza { return WithBox{Base: p} }, " + box", 30},
	} {
		base := &pizzaProbe{name: "Special", price: 417}
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
	base := &pizzaProbe{name: "Test", price: 400}
	p := WithCheese{Base: WithBox{Base: base}}
	if p.Name() != "Test + box + cheese" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithCheese{Base: WithCheese{Base: base}}
	if double.Name() != "Test + cheese + cheese" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if Pickup(base) != base || Courier(base) != base+100 || Express(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   DeliveryPrice
	}{
		{"pickup", Pickup}, {"courier", Courier}, {"express", Express},
	} {
		for _, kind := range []string{"margherita", "pepperoni"} {
			for _, cheese := range []bool{false, true} {
				for _, box := range []bool{false, true} {
					_, base, _ := legacyOrder(kind, cheese, box, "pickup")
					_, want, _ := legacyOrder(kind, cheese, box, tc.name)
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
