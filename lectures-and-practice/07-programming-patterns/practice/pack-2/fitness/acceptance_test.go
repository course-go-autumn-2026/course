package fitness

import (
	"errors"
	"reflect"
	"testing"
)

func TestAcceptanceValidationRegression(t *testing.T) {
	for _, tc := range []struct {
		o    Visit
		want error
	}{
		{Visit{}, ErrEmpty}, {Visit{Attendees: -1, Total: 100}, ErrEmpty},
		{Visit{Attendees: 1}, ErrTotal}, {Visit{Attendees: 1, Total: -1}, ErrTotal},
		{Visit{Attendees: 6}, ErrTotal}, {Visit{Attendees: 6, Total: 1}, ErrApproval},
		{Visit{Attendees: 5, Total: 1}, nil}, {Visit{Attendees: 1, Total: 1}, nil},
	} {
		if err := Validate(tc.o); !errors.Is(err, tc.want) || !errors.Is(err, legacyValidate(tc.o)) {
			t.Fatal(tc.o, err, tc.want)
		}
	}
}
func TestAcceptanceIndependentChecks(t *testing.T) {
	for _, tc := range []struct {
		check Check
		o     Visit
		want  error
	}{
		{CheckAttendees, Visit{}, ErrEmpty}, {CheckAttendees, Visit{Attendees: 1}, nil},
		{CheckTotal, Visit{}, ErrTotal}, {CheckTotal, Visit{Total: 1}, nil},
		{CheckLimit, Visit{Attendees: 6}, ErrApproval}, {CheckLimit, Visit{Attendees: 5}, nil},
	} {
		if err := tc.check(tc.o); !errors.Is(err, tc.want) {
			t.Fatal(err, tc.want)
		}
	}
}
func TestAcceptanceChainOrderAndStop(t *testing.T) {
	want := Visit{ID: 42, Attendees: 2, Total: 900}
	cause := errors.New("dependency offline")
	calls := []int{}
	step := func(n int, err error) Check {
		return func(got Visit) error {
			if got != want {
				t.Fatal(got)
			}
			calls = append(calls, n)
			return err
		}
	}
	if err := Chain(step(1, nil), step(2, cause), step(3, nil))(want); !errors.Is(err, cause) {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, []int{1, 2}) {
		t.Fatal(calls)
	}
}
func TestAcceptanceChainEmptyAndSuccess(t *testing.T) {
	if err := Chain()(Visit{}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	step := func(Visit) error { calls++; return nil }
	if err := Chain(step, step)(Visit{}); err != nil || calls != 2 {
		t.Fatal(err, calls)
	}
}
func TestAcceptanceObserverOrderAndEvent(t *testing.T) {
	n := &Notifier{}
	calls := []int{}
	want := Event{VisitID: 42, Status: "registered"}
	for i := 1; i <= 3; i++ {
		id := i
		n.Subscribe(ObserverFunc(func(got Event) error {
			if got != want {
				t.Fatal(got)
			}
			calls = append(calls, id)
			return nil
		}))
	}
	if len(calls) != 0 {
		t.Fatal("Subscribe emitted an event")
	}
	if err := n.Publish(want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, []int{1, 2, 3}) {
		t.Fatal(calls)
	}
}
func TestAcceptanceObserverErrorsAndContinuation(t *testing.T) {
	a, b := errors.New("primary subscriber offline"), errors.New("secondary subscriber offline")
	calls := []int{}
	n := &Notifier{}
	n.Subscribe(ObserverFunc(func(Event) error { calls = append(calls, 1); return a }))
	n.Subscribe(ObserverFunc(func(Event) error { calls = append(calls, 2); return nil }))
	n.Subscribe(ObserverFunc(func(Event) error { calls = append(calls, 3); return b }))
	err := n.Publish(Event{VisitID: 7, Status: "registered"})
	if !errors.Is(err, a) || !errors.Is(err, b) || !reflect.DeepEqual(calls, []int{1, 2, 3}) {
		t.Fatal(err, calls)
	}
}
func TestAcceptanceObserverEmptyDuplicateAndRepeat(t *testing.T) {
	n := &Notifier{}
	if err := n.Publish(Event{}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	observer := ObserverFunc(func(Event) error { calls++; return nil })
	n.Subscribe(observer)
	n.Subscribe(observer)
	for i := 0; i < 2; i++ {
		if err := n.Publish(Event{VisitID: int64(i)}); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 4 {
		t.Fatal(calls)
	}
}
func TestAcceptanceObserverLegacyEvent(t *testing.T) {
	o := Visit{ID: 9}
	want := []Event{}
	old := ObserverFunc(func(e Event) error { want = append(want, e); return nil })
	if err := legacyRegistered(o, old, old); err != nil {
		t.Fatal(err)
	}
	got := []Event{}
	capture := ObserverFunc(func(e Event) error { got = append(got, e); return nil })
	n := &Notifier{}
	n.Subscribe(capture)
	n.Subscribe(capture)
	if err := n.Registered(o); err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal(err, got, want)
	}
}
func TestAcceptanceOptionsDefaultsAndIsolation(t *testing.T) {
	custom, err := NewClub(WithMaxAttendees(8), WithRegistrationFee(0))
	if err != nil || custom == nil {
		t.Fatal(err)
	}
	plain, err := NewClub()
	if err != nil || plain == nil {
		t.Fatal(err)
	}
	if custom.Config() != (Config{MaxAttendees: 8, RegistrationFee: 0}) || plain.Config() != (Config{MaxAttendees: 5, RegistrationFee: 100}) {
		t.Fatal(custom.Config(), plain.Config())
	}
	if !custom.CanHandle(8) || custom.CanHandle(9) || custom.TotalWithFee(450) != 450 || plain.TotalWithFee(450) != 550 {
		t.Fatal("settings ignored")
	}
}
func TestAcceptanceOptionsLegacyRegression(t *testing.T) {
	for _, max := range []int{1, 5, 8} {
		for _, fee := range []int{0, 100, 250} {
			want, _ := legacyNewClub(max, fee)
			got, err := NewClub(WithMaxAttendees(max), WithRegistrationFee(fee))
			if err != nil || got == nil {
				t.Fatal(err)
			}
			if got.Config() != want.Config() {
				t.Fatal(got.Config(), want.Config())
			}
		}
	}
}
func TestAcceptanceOptionsOrderAndExtension(t *testing.T) {
	calls := []int{}
	first := func(c *Config) error {
		calls = append(calls, 1)
		if *c != (Config{MaxAttendees: 5, RegistrationFee: 100}) {
			t.Fatal(*c)
		}
		c.RegistrationFee += 10
		return nil
	}
	second := func(c *Config) error { calls = append(calls, 2); c.RegistrationFee *= 2; return nil }
	p, err := NewClub(first, second)
	if err != nil || p == nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, []int{1, 2}) || p.Config().RegistrationFee != 220 {
		t.Fatal(calls, p.Config())
	}
	p, err = NewClub(WithMaxAttendees(8), WithMaxAttendees(3))
	if err != nil || p == nil {
		t.Fatal(err)
	}
	if p.Config().MaxAttendees != 3 {
		t.Fatal(p.Config())
	}
}
func TestAcceptanceOptionsInvalidAndStop(t *testing.T) {
	for _, tc := range []struct {
		option Option
		cause  error
	}{
		{WithMaxAttendees(0), ErrInvalidMax}, {WithMaxAttendees(-1), ErrInvalidMax}, {WithRegistrationFee(-1), ErrInvalidFee},
	} {
		p, err := NewClub(tc.option)
		if p != nil || !errors.Is(err, tc.cause) {
			t.Fatal(p, err, tc.cause)
		}
	}
	cause := errors.New("configuration unavailable")
	later := false
	p, err := NewClub(func(*Config) error { return cause }, func(*Config) error { later = true; return nil })
	if p != nil || !errors.Is(err, cause) || later {
		t.Fatal(p, err, later)
	}
}
