package autoservice

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
		p     Service
		name  string
		price int
	}{
		{Inspection{}, "Inspection", 450}, {WheelAlignment{}, "WheelAlignment", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"inspection", "alignment"} {
		name, price, _ := legacyService(kind, false, false, "scheduled")
		got, err := NewService(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "Inspection"} {
		p, err := NewService(kind)
		if p != nil || !errors.Is(err, ErrUnknownService) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"inspection", "alignment"} {
		for _, carWash := range []bool{false, true} {
			for _, prioritySlot := range []bool{false, true} {
				baseName, basePrice, _ := legacyService(kind, false, false, "scheduled")
				var p Service = &productProbe{name: baseName, price: basePrice}
				if carWash {
					p = WithCarWash{Base: p}
				}
				if prioritySlot {
					p = WithPrioritySlot{Base: p}
				}
				name, price, _ := legacyService(kind, carWash, prioritySlot, "scheduled")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, carWash, prioritySlot, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(Service) Service
		suffix string
		extra  int
	}{
		{func(p Service) Service { return WithCarWash{Base: p} }, " + car wash", 50},
		{func(p Service) Service { return WithPrioritySlot{Base: p} }, " + priority slot", 30},
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
	p := WithCarWash{Base: WithPrioritySlot{Base: base}}
	if p.Name() != "Test + priority slot + car wash" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithCarWash{Base: WithCarWash{Base: base}}
	if double.Name() != "Test + car wash + car wash" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if Scheduled(base) != base || SameDay(base) != base+100 || NightShift(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   ServicePrice
	}{
		{"scheduled", Scheduled}, {"same_day", SameDay}, {"night_shift", NightShift},
	} {
		for _, kind := range []string{"inspection", "alignment"} {
			for _, carWash := range []bool{false, true} {
				for _, prioritySlot := range []bool{false, true} {
					_, base, _ := legacyService(kind, carWash, prioritySlot, "scheduled")
					_, want, _ := legacyService(kind, carWash, prioritySlot, tc.name)
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
