package domainerror_test

import (
	"errors"
	"testing"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/domainerror"
)

func TestProblemMapsDomainErrors(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "validation", err: &trip.ValidationError{Field: "price", Message: "cannot be negative"}, wantStatus: 400, wantCode: "invalid_request"},
		{name: "not found", err: trip.ErrTripNotFound, wantStatus: 404, wantCode: "trip_not_found"},
		{name: "driver busy", err: trip.ErrDriverBusy, wantStatus: 409, wantCode: "driver_busy"},
		{name: "completed", err: trip.ErrTripCompleted, wantStatus: 409, wantCode: "trip_completed"},
		{name: "internal", err: errors.New("database unavailable"), wantStatus: 500, wantCode: "internal_error"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			status, body := domainerror.Problem(testCase.err, "/api/v1/trips")
			if status != testCase.wantStatus {
				t.Errorf("status = %d, want %d", status, testCase.wantStatus)
			}
			if body.Code != testCase.wantCode {
				t.Errorf("code = %q, want %q", body.Code, testCase.wantCode)
			}
			if status == 500 && body.Detail != nil && *body.Detail == "database unavailable" {
				t.Error("internal error detail leaked")
			}
		})
	}
}
