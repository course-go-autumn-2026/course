package courses

import (
	"errors"
	"reflect"
	"testing"
)

func TestLegacyValidation(t *testing.T) {
	if !errors.Is(legacyValidate(Enrollment{}), ErrEmpty) {
		t.Fatal("empty")
	}
	if !errors.Is(legacyValidate(Enrollment{Students: 6}), ErrTotal) {
		t.Fatal("priority")
	}
	if !errors.Is(legacyValidate(Enrollment{Students: 6, Total: 1}), ErrApproval) {
		t.Fatal("limit")
	}
	if err := legacyValidate(Enrollment{Students: 5, Total: 1}); err != nil {
		t.Fatal(err)
	}
}
func TestLegacyNotification(t *testing.T) {
	a, b := errors.New("operator"), errors.New("client")
	events := []Event{}
	one := ObserverFunc(func(e Event) error { events = append(events, e); return a })
	two := ObserverFunc(func(e Event) error { events = append(events, e); return b })
	err := legacyActivated(Enrollment{ID: 42}, one, two)
	if !errors.Is(err, a) || !errors.Is(err, b) {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(events, []Event{{EnrollmentID: 42, Status: "activated"}, {EnrollmentID: 42, Status: "activated"}}) {
		t.Fatal(events)
	}
}
func TestLegacyConfiguration(t *testing.T) {
	p, err := legacyNewSchool(5, 100)
	if err != nil || p == nil {
		t.Fatal(err)
	}
	if p.Config() != (Config{MaxStudents: 5, EnrollmentFee: 100}) {
		t.Fatal(p.Config())
	}
	for _, tc := range []struct {
		max, fee int
		want     error
	}{{0, 100, ErrInvalidMax}, {5, -1, ErrInvalidFee}} {
		p, err := legacyNewSchool(tc.max, tc.fee)
		if p != nil || !errors.Is(err, tc.want) {
			t.Fatal(p, err)
		}
	}
}
