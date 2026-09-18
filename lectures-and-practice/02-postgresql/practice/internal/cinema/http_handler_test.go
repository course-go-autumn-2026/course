package cinema

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseScreeningFilter(t *testing.T) {
	request := httptest.NewRequest(
		"GET",
		"/screenings?search=%D0%B8%D0%BD&starts_from=2026-09-10T18%3A00%3A00Z&min_seats=2&limit=10",
		nil,
	)

	filter, err := parseScreeningFilter(request)
	if err != nil {
		t.Fatalf("parseScreeningFilter() error = %v", err)
	}
	if filter.Search != "ин" {
		t.Errorf("Search = %q, want %q", filter.Search, "ин")
	}
	wantStartsFrom := time.Date(2026, 9, 10, 18, 0, 0, 0, time.UTC)
	if filter.StartsFrom == nil || !filter.StartsFrom.Equal(wantStartsFrom) {
		t.Errorf("StartsFrom = %v, want %v", filter.StartsFrom, wantStartsFrom)
	}
	if filter.MinSeats == nil || *filter.MinSeats != 2 {
		t.Errorf("MinSeats = %v, want 2", filter.MinSeats)
	}
	if filter.Limit != 10 {
		t.Errorf("Limit = %d, want 10", filter.Limit)
	}
}

func TestParseScreeningFilterDefaults(t *testing.T) {
	request := httptest.NewRequest("GET", "/screenings", nil)

	filter, err := parseScreeningFilter(request)
	if err != nil {
		t.Fatalf("parseScreeningFilter() error = %v", err)
	}
	if filter.Limit != 20 {
		t.Errorf("Limit = %d, want 20", filter.Limit)
	}
}

func TestParseScreeningFilterRejectsInvalidValues(t *testing.T) {
	tests := []string{
		"/screenings?starts_from=tomorrow",
		"/screenings?min_seats=-1",
		"/screenings?limit=0",
		"/screenings?limit=101",
	}

	for _, target := range tests {
		t.Run(target, func(t *testing.T) {
			request := httptest.NewRequest("GET", target, nil)
			if _, err := parseScreeningFilter(request); err == nil {
				t.Fatal("parseScreeningFilter() error = nil, want an error")
			}
		})
	}
}
