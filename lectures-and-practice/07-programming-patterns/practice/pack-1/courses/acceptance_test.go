package courses

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
		p     Course
		name  string
		price int
	}{
		{RecordedCourse{}, "RecordedCourse", 450}, {LiveCourse{}, "LiveCourse", 550},
	} {
		if tc.p.Name() != tc.name || tc.p.Price() != tc.price {
			t.Fatal(tc.p.Name(), tc.p.Price())
		}
	}
}
func TestAcceptanceFactoryRegression(t *testing.T) {
	for _, kind := range []string{"recorded", "live"} {
		name, price, _ := legacyEnrollment(kind, false, false, "self")
		got, err := NewCourse(kind)
		if err != nil || got == nil {
			t.Fatal(kind, err)
		}
		if got.Name() != name || got.Price() != price {
			t.Fatal(kind, got.Name(), got.Price())
		}
	}
}
func TestAcceptanceFactoryUnknown(t *testing.T) {
	for _, kind := range []string{"", "unknown", "RecordedCourse"} {
		p, err := NewCourse(kind)
		if p != nil || !errors.Is(err, ErrUnknownCourse) {
			t.Fatal(kind, p, err)
		}
	}
}
func TestAcceptanceDecoratorRegression(t *testing.T) {
	for _, kind := range []string{"recorded", "live"} {
		for _, mentoring := range []bool{false, true} {
			for _, certificate := range []bool{false, true} {
				baseName, basePrice, _ := legacyEnrollment(kind, false, false, "self")
				var p Course = &productProbe{name: baseName, price: basePrice}
				if mentoring {
					p = WithMentoring{Base: p}
				}
				if certificate {
					p = WithCertificate{Base: p}
				}
				name, price, _ := legacyEnrollment(kind, mentoring, certificate, "self")
				if p.Name() != name || p.Price() != price {
					t.Fatal(kind, mentoring, certificate, p.Name(), p.Price())
				}
			}
		}
	}
}
func TestAcceptanceDecoratorDelegation(t *testing.T) {
	for _, tc := range []struct {
		wrap   func(Course) Course
		suffix string
		extra  int
	}{
		{func(p Course) Course { return WithMentoring{Base: p} }, " + mentoring", 50},
		{func(p Course) Course { return WithCertificate{Base: p} }, " + certificate", 30},
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
	p := WithMentoring{Base: WithCertificate{Base: base}}
	if p.Name() != "Test + certificate + mentoring" || p.Price() != 480 {
		t.Fatal(p.Name(), p.Price())
	}
	double := WithMentoring{Base: WithMentoring{Base: base}}
	if double.Name() != "Test + mentoring + mentoring" || double.Price() != 500 {
		t.Fatal(double.Name(), double.Price())
	}
}
func TestAcceptanceStrategyTariffs(t *testing.T) {
	for _, base := range []int{0, 50, 450, 1234} {
		if SelfEnrollment(base) != base || AssistedEnrollment(base) != base+100 || PriorityEnrollment(base) != base+200 {
			t.Fatal(base)
		}
	}
}
func TestAcceptanceStrategyRegression(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   EnrollmentPrice
	}{
		{"self", SelfEnrollment}, {"assisted", AssistedEnrollment}, {"priority", PriorityEnrollment},
	} {
		for _, kind := range []string{"recorded", "live"} {
			for _, mentoring := range []bool{false, true} {
				for _, certificate := range []bool{false, true} {
					_, base, _ := legacyEnrollment(kind, mentoring, certificate, "self")
					_, want, _ := legacyEnrollment(kind, mentoring, certificate, tc.name)
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
