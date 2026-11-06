package domainerror_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/domainerror"
)

func TestProblem(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"not found", fmt.Errorf("fetch: %w", trip.ErrNotFound), 404, "trip_not_found"},
		{"completed", trip.ErrCompleted, 409, "trip_completed"},
		{"driver busy", trip.ErrDriverBusy, 409, "driver_busy"},
		{"key conflict", trip.ErrIdempotencyConflict, 409, "idempotency_conflict"},
		{"request", domainerror.ErrInvalidRequest, 400, "invalid_request"},
		{"unknown route", domainerror.ErrRouteNotFound, 404, "route_not_found"},
		{"unsupported method", domainerror.ErrMethodNotAllowed, 405, "method_not_allowed"},
		{"domain validation", fmt.Errorf("create: %w", &trip.ValidationError{Field: "price", Reason: "negative"}), 400, "invalid_request"},
		{"internal", errors.New("private database credentials"), 500, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := domainerror.Problem(tc.err)
			if status != tc.status || int(body.Status) != tc.status || body.Code != tc.code {
				t.Fatalf("status %d, body %+v", status, body)
			}
			if body.Detail == nil || body.Title == "" || body.Type == "" {
				t.Fatalf("incomplete Problem: %+v", body)
			}
			if strings.Contains(*body.Detail, "private") {
				t.Fatalf("internal details leaked: %+v", body)
			}
		})
	}
}
