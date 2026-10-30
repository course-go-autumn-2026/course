package library

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
		p     LoanPlan
		name  string
		price int
	}{
		{StandardLoan{}, "StandardLoan", 450}, {ExtendedLoan{}, "ExtendedLoan", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"standard", "extended"} {
		name, price, _ := legacyLoan(kind, false, false, "desk_pickup")
		got, err := NewLoanPlan(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "StandardLoan"} {
		p, err := NewLoanPlan(kind)
		if p != nil || !errors.Is(err, ErrUnknownLoanPlan) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"standard", "extended"} {
		for _, scanning := range []bool{false, true} {
			for _, protectiveCover := range []bool{false, true} {
				baseName, basePrice, _ := legacyLoan(kind, false, false, "desk_pickup")
				var p LoanPlan = &productProbe{name: baseName, price: basePrice}
				if scanning {
					p = WithScanning{Base: p}
				}
				if protectiveCover {
					p = WithProtectiveCover{Base: p}
				}
				name, price, _ := legacyLoan(kind, scanning, protectiveCover, "desk_pickup")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, scanning, protectiveCover, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(LoanPlan) LoanPlan
		suffix string
		extra  int
	}{
		{func(p LoanPlan) LoanPlan { return WithScanning{Base: p} }, " + scanning", 50},
		{func(p LoanPlan) LoanPlan { return WithProtectiveCover{Base: p} }, " + protective cover", 30},
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
	p := WithScanning{Base: WithProtectiveCover{Base: base}}
	if p.Name() != "Test + protective cover + scanning" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithScanning{Base: WithScanning{Base: base}}
	if double.Name() != "Test + scanning + scanning" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if DeskPickup(base) != base || HomeDelivery(base) != base+100 || PriorityDelivery(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   LoanPrice
	}{
		{"desk_pickup", DeskPickup}, {"home_delivery", HomeDelivery}, {"priority_delivery", PriorityDelivery},
	} {
		for _, kind := range []string{"standard", "extended"} {
			for _, scanning := range []bool{false, true} {
				for _, protectiveCover := range []bool{false, true} {
					_, base, _ := legacyLoan(kind, scanning, protectiveCover, "desk_pickup")
					_, want, _ := legacyLoan(kind, scanning, protectiveCover, tc.name)
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
