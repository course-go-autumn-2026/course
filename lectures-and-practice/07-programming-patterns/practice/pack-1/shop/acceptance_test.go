package shop

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
		p     Product
		name  string
		price int
	}{
		{Keyboard{}, "Keyboard", 450}, {Mouse{}, "Mouse", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"keyboard", "mouse"} {
		name, price, _ := legacyPurchase(kind, false, false, "store_pickup")
		got, err := NewProduct(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "Keyboard"} {
		p, err := NewProduct(kind)
		if p != nil || !errors.Is(err, ErrUnknownProduct) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"keyboard", "mouse"} {
		for _, warranty := range []bool{false, true} {
			for _, giftWrap := range []bool{false, true} {
				baseName, basePrice, _ := legacyPurchase(kind, false, false, "store_pickup")
				var p Product = &productProbe{name: baseName, price: basePrice}
				if warranty {
					p = WithWarranty{Base: p}
				}
				if giftWrap {
					p = WithGiftWrap{Base: p}
				}
				name, price, _ := legacyPurchase(kind, warranty, giftWrap, "store_pickup")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, warranty, giftWrap, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(Product) Product
		suffix string
		extra  int
	}{
		{func(p Product) Product { return WithWarranty{Base: p} }, " + warranty", 50},
		{func(p Product) Product { return WithGiftWrap{Base: p} }, " + gift wrap", 30},
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
	p := WithWarranty{Base: WithGiftWrap{Base: base}}
	if p.Name() != "Test + gift wrap + warranty" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithWarranty{Base: WithWarranty{Base: base}}
	if double.Name() != "Test + warranty + warranty" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if StorePickup(base) != base || Courier(base) != base+100 || Express(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   FulfillmentPrice
	}{
		{"store_pickup", StorePickup}, {"courier", Courier}, {"express", Express},
	} {
		for _, kind := range []string{"keyboard", "mouse"} {
			for _, warranty := range []bool{false, true} {
				for _, giftWrap := range []bool{false, true} {
					_, base, _ := legacyPurchase(kind, warranty, giftWrap, "store_pickup")
					_, want, _ := legacyPurchase(kind, warranty, giftWrap, tc.name)
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
