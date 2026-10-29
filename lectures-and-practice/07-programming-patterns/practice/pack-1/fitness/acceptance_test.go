package fitness

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
		p     Membership
		name  string
		price int
	}{
		{DaytimePass{}, "DaytimePass", 450}, {FullDayPass{}, "FullDayPass", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"daytime", "full_day"} {
		name, price, _ := legacyMembership(kind, false, false, "next_week")
		got, err := NewMembership(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "DaytimePass"} {
		p, err := NewMembership(kind)
		if p != nil || !errors.Is(err, ErrUnknownMembership) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"daytime", "full_day"} {
		for _, locker := range []bool{false, true} {
			for _, towel := range []bool{false, true} {
				baseName, basePrice, _ := legacyMembership(kind, false, false, "next_week")
				var p Membership = &productProbe{name: baseName, price: basePrice}
				if locker {
					p = WithLocker{Base: p}
				}
				if towel {
					p = WithTowel{Base: p}
				}
				name, price, _ := legacyMembership(kind, locker, towel, "next_week")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, locker, towel, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(Membership) Membership
		suffix string
		extra  int
	}{
		{func(p Membership) Membership { return WithLocker{Base: p} }, " + locker", 50},
		{func(p Membership) Membership { return WithTowel{Base: p} }, " + towel", 30},
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
	p := WithLocker{Base: WithTowel{Base: base}}
	if p.Name() != "Test + towel + locker" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithLocker{Base: WithLocker{Base: base}}
	if double.Name() != "Test + locker + locker" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if StartNextWeek(base) != base || StartToday(base) != base+100 || PersonalOnboarding(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   ActivationPrice
	}{
		{"next_week", StartNextWeek}, {"today", StartToday}, {"personal", PersonalOnboarding},
	} {
		for _, kind := range []string{"daytime", "full_day"} {
			for _, locker := range []bool{false, true} {
				for _, towel := range []bool{false, true} {
					_, base, _ := legacyMembership(kind, locker, towel, "next_week")
					_, want, _ := legacyMembership(kind, locker, towel, tc.name)
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
