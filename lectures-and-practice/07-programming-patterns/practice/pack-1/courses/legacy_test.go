package courses

import (
	"errors"
	"testing"
)

func TestLegacyMenu(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		price      int
	}{
		{"recorded", "RecordedCourse", 450}, {"live", "LiveCourse", 550},
	} {
		name, price, err := legacyEnrollment(tc.kind, false, false, "self")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyEnrollment("unknown", false, false, "self"); name != "" || price != 0 || !errors.Is(err, ErrUnknownCourse) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyExtras(t *testing.T) {
	for _, tc := range []struct {
		mentoring, certificate bool
		name                   string
		price                  int
	}{
		{false, false, "RecordedCourse", 450}, {true, false, "RecordedCourse + mentoring", 500},
		{false, true, "RecordedCourse + certificate", 480}, {true, true, "RecordedCourse + mentoring + certificate", 530},
	} {
		name, price, err := legacyEnrollment("recorded", tc.mentoring, tc.certificate, "self")
		if err != nil || name != tc.name || price != tc.price {
			t.Fatal(name, price, err)
		}
	}
	if name, price, err := legacyEnrollment("unknown", true, true, "self"); name != "" || price != 0 || !errors.Is(err, ErrUnknownCourse) {
		t.Fatal(name, price, err)
	}
}
func TestLegacyTariff(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"self", 450}, {"assisted", 550}, {"priority", 650}} {
		if _, got, err := legacyEnrollment("recorded", false, false, tc.name); got != tc.want || err != nil {
			t.Fatal(got, err)
		}
	}
	if name, got, err := legacyEnrollment("recorded", false, false, "unknown"); name != "" || got != 0 || !errors.Is(err, ErrUnknownTariff) {
		t.Fatal(name, got, err)
	}
}
