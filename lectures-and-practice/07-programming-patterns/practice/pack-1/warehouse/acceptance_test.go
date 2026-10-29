package warehouse

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
		p     StorageCell
		name  string
		price int
	}{
		{SmallCell{}, "SmallCell", 450}, {LargeCell{}, "LargeCell", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"small", "large"} {
		name, price, _ := legacyStorage(kind, false, false, "self")
		got, err := NewStorageCell(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "SmallCell"} {
		p, err := NewStorageCell(kind)
		if p != nil || !errors.Is(err, ErrUnknownStorageCell) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"small", "large"} {
		for _, climateControl := range []bool{false, true} {
			for _, insurance := range []bool{false, true} {
				baseName, basePrice, _ := legacyStorage(kind, false, false, "self")
				var p StorageCell = &productProbe{name: baseName, price: basePrice}
				if climateControl {
					p = WithClimateControl{Base: p}
				}
				if insurance {
					p = WithInsurance{Base: p}
				}
				name, price, _ := legacyStorage(kind, climateControl, insurance, "self")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, climateControl, insurance, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(StorageCell) StorageCell
		suffix string
		extra  int
	}{
		{func(p StorageCell) StorageCell { return WithClimateControl{Base: p} }, " + climate control", 50},
		{func(p StorageCell) StorageCell { return WithInsurance{Base: p} }, " + insurance", 30},
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
	p := WithClimateControl{Base: WithInsurance{Base: base}}
	if p.Name() != "Test + insurance + climate control" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithClimateControl{Base: WithClimateControl{Base: base}}
	if double.Name() != "Test + climate control + climate control" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if SelfHandling(base) != base || StandardHandling(base) != base+100 || ExpressHandling(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   HandlingPrice
	}{
		{"self", SelfHandling}, {"standard", StandardHandling}, {"express", ExpressHandling},
	} {
		for _, kind := range []string{"small", "large"} {
			for _, climateControl := range []bool{false, true} {
				for _, insurance := range []bool{false, true} {
					_, base, _ := legacyStorage(kind, climateControl, insurance, "self")
					_, want, _ := legacyStorage(kind, climateControl, insurance, tc.name)
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
