package pizzeria

import (
	"errors"
	"reflect"
	"testing"
)

func TestLegacyValidation(t *testing.T) {
	if !errors.Is(legacyValidate(Order{}), ErrEmpty) {
		t.Fatal("empty")
	}
	if !errors.Is(legacyValidate(Order{Pizzas: 6}), ErrTotal) {
		t.Fatal("priority")
	}
	if !errors.Is(legacyValidate(Order{Pizzas: 6, Total: 1}), ErrApproval) {
		t.Fatal("limit")
	}
	if err := legacyValidate(Order{Pizzas: 5, Total: 1}); err != nil {
		t.Fatal(err)
	}
}
func TestLegacyNotification(t *testing.T) {
	a, b := errors.New("screen"), errors.New("customer")
	events := []Event{}
	one := ObserverFunc(func(e Event) error { events = append(events, e); return a })
	two := ObserverFunc(func(e Event) error { events = append(events, e); return b })
	err := legacyReady(Order{ID: 42}, one, two)
	if !errors.Is(err, a) || !errors.Is(err, b) {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(events, []Event{{OrderID: 42, Status: "ready"}, {OrderID: 42, Status: "ready"}}) {
		t.Fatal(events)
	}
}
func TestLegacyConfiguration(t *testing.T) {
	p, err := legacyNewPizzeria(5, 100)
	if err != nil || p == nil {
		t.Fatal(err)
	}
	if p.Config() != (Config{MaxPizzas: 5, DeliveryFee: 100}) {
		t.Fatal(p.Config())
	}
	for _, tc := range []struct {
		max, fee int
		want     error
	}{{0, 100, ErrInvalidMax}, {5, -1, ErrInvalidFee}} {
		p, err := legacyNewPizzeria(tc.max, tc.fee)
		if p != nil || !errors.Is(err, tc.want) {
			t.Fatal(p, err)
		}
	}
}
