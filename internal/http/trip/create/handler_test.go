package create_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/trip/create"
	"github.com/DaleCoopTP/grishin_lab1/internal/model"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestHandler(t *testing.T) {
	id, user, driver, key := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	input := model.CreateTripInput{UserID: user, DriverID: driver,
		StartPoint: trip.Point{Latitude: 59.9, Longitude: 30.3},
		EndPoint:   trip.Point{Latitude: 60, Longitude: 31}, Price: 1450}
	requestBody := fmt.Sprintf(`{"user_id":%q,"driver_id":%q,"start_point":{"latitude":59.9,"longitude":30.3},"end_point":{"latitude":60,"longitude":31},"price":1450}`, user, driver)
	cases := []struct {
		name     string
		replayed bool
		err      error
		status   int
		code     string
	}{
		{"created", false, nil, 201, ""},
		{"replayed", true, nil, 200, ""},
		{"busy", false, trip.ErrDriverBusy, 409, "driver_busy"},
		{"conflict", false, trip.ErrIdempotencyConflict, 409, "idempotency_conflict"},
		{"internal", false, errors.New("private database details"), 500, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			usecase := NewMockusecase(gomock.NewController(t))
			result := trip.Trip{ID: id, UserID: user, DriverID: driver,
				StartPoint: input.StartPoint, EndPoint: input.EndPoint, Price: input.Price,
				Status: trip.StatusActive, StartedAt: time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC)}
			usecase.EXPECT().CreateWithKey(gomock.Any(), input, &key).Return(result, tc.replayed, tc.err)
			writer := response.New(slog.New(slog.NewTextHandler(io.Discard, nil)))
			handler := create.New(usecase, writer)
			req := httptest.NewRequest("POST", "/api/v1/trips", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			out := httptest.NewRecorder()
			handler.Handle(out, req, api.CreateTripParams{IdempotencyKey: &key})
			if out.Code != tc.status {
				t.Fatalf("status %d, body %s", out.Code, out.Body.String())
			}
			if tc.err == nil {
				var body api.Trip
				if err := json.Unmarshal(out.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Id != id || body.UserId != user || body.DriverId != driver || body.Price != input.Price ||
					body.StartPoint.Latitude != input.StartPoint.Latitude || body.EndPoint.Longitude != input.EndPoint.Longitude ||
					body.Status != api.Active || out.Header().Get("Location") != "/api/v1/trips/"+id.String() ||
					out.Header().Get("Content-Type") != "application/json" {
					t.Fatalf("unexpected created trip: %+v", body)
				}
			} else {
				assertProblem(t, out, tc.code)
			}
		})
	}
}

func TestHandlerRejectsInvalidRequest(t *testing.T) {
	user, driver := uuid.New(), uuid.New()
	valid := fmt.Sprintf(`{"user_id":%q,"driver_id":%q,"start_point":{"latitude":0,"longitude":0},"end_point":{"latitude":0,"longitude":0},"price":0}`, user, driver)
	cases := []struct{ name, body, contentType string }{
		{"malformed JSON", `{`, "application/json"},
		{"missing fields", `{}`, "application/json"},
		{"null field", strings.Replace(valid, `"price":0`, `"price":null`, 1), "application/json"},
		{"negative price", strings.Replace(valid, `"price":0`, `"price":-1`, 1), "application/json"},
		{"latitude", strings.Replace(valid, `"latitude":0`, `"latitude":91`, 1), "application/json"},
		{"unknown field", strings.Replace(valid, `"price":0`, `"price":0,"extra":true`, 1), "application/json"},
		{"zero UUID", strings.Replace(valid, user.String(), uuid.Nil.String(), 1), "application/json"},
		{"invalid UUID", strings.Replace(valid, user.String(), "bad", 1), "application/json"},
		{"multiple objects", valid + `{}`, "application/json"},
		{"body too large", strings.Repeat(" ", (1<<20)+1), "application/json"},
		{"content type", valid, "text/plain"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			usecase := NewMockusecase(gomock.NewController(t))
			writer := response.New(slog.New(slog.NewTextHandler(io.Discard, nil)))
			handler := create.New(usecase, writer)
			req := httptest.NewRequest("POST", "/api/v1/trips", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			out := httptest.NewRecorder()
			handler.Handle(out, req, api.CreateTripParams{})
			if out.Code != 400 {
				t.Fatalf("status %d, body %s", out.Code, out.Body.String())
			}
			assertProblem(t, out, "invalid_request")
		})
	}
}

func assertProblem(t *testing.T, out *httptest.ResponseRecorder, code string) {
	t.Helper()
	var body api.Problem
	if err := json.Unmarshal(out.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != code || body.Status != int32(out.Code) || body.Instance == nil ||
		*body.Instance != "/api/v1/trips" || out.Header().Get("Content-Type") != "application/problem+json" ||
		strings.Contains(out.Body.String(), "private") {
		t.Fatalf("unexpected Problem: %s", out.Body.String())
	}
}
