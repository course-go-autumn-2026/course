package httptransport_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	httptransport "github.com/DaleCoopTP/grishin_lab1/internal/http"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestRouter(t *testing.T) {
	id, key := uuid.New(), uuid.New()
	cases := []struct {
		name, method, path string
		invalid            bool
		key                string
	}{
		{"create", "POST", "/api/v1/trips", false, key.String()},
		{"get", "GET", "/api/v1/trips/" + id.String(), false, ""},
		{"finish", "POST", "/api/v1/trips/" + id.String() + "/finish", false, ""},
		{"health", "GET", "/health", false, ""},
		{"ready", "GET", "/ready", false, ""},
		{"invalid get UUID", "GET", "/api/v1/trips/bad", true, ""},
		{"invalid finish UUID", "POST", "/api/v1/trips/bad/finish", true, ""},
		{"invalid key", "POST", "/api/v1/trips", true, "bad"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			create := NewMockcreateHandler(ctrl)
			get, finish := NewMocktripHandler(ctrl), NewMocktripHandler(ctrl)
			health, ready := NewMockHandler(ctrl), NewMockHandler(ctrl)
			switch tc.name {
			case "create":
				create.EXPECT().Handle(gomock.Any(), gomock.Any(), api.CreateTripParams{IdempotencyKey: &key}).Do(
					func(w http.ResponseWriter, _ *http.Request, _ api.CreateTripParams) { w.WriteHeader(204) })
			case "get":
				get.EXPECT().Handle(gomock.Any(), gomock.Any(), id).Do(
					func(w http.ResponseWriter, _ *http.Request, _ api.TripId) { w.WriteHeader(204) })
			case "finish":
				finish.EXPECT().Handle(gomock.Any(), gomock.Any(), id).Do(
					func(w http.ResponseWriter, _ *http.Request, _ api.TripId) { w.WriteHeader(204) })
			case "health":
				health.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Do(
					func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })
			case "ready":
				ready.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).Do(
					func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })
			}
			writer := response.New(slog.New(slog.NewTextHandler(io.Discard, nil)))
			router := httptransport.NewRouter(create, get, finish, health, ready, writer)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.key != "" {
				req.Header.Set("Idempotency-Key", tc.key)
			}
			out := httptest.NewRecorder()
			router.ServeHTTP(out, req)
			expected := 204
			if tc.invalid {
				expected = 400
				var body api.Problem
				if err := json.Unmarshal(out.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Code != "invalid_request" || out.Header().Get("Content-Type") != "application/problem+json" {
					t.Fatalf("unexpected binding error: %s", out.Body.String())
				}
			}
			if out.Code != expected {
				t.Fatalf("status %d, body %s", out.Code, out.Body.String())
			}
		})
	}
}
