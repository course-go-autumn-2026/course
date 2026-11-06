package httptransport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	httptransport "github.com/DaleCoopTP/grishin_lab1/internal/http"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
)

const tripsPath = "/api/v1/trips"

func TestRouter(t *testing.T) {
	id, key := uuid.New(), uuid.New()
	cases := []struct {
		name, method, path string
		invalid            bool
		key                string
	}{
		{"create", http.MethodPost, tripsPath, false, key.String()},
		{"get", http.MethodGet, tripsPath + "/" + id.String(), false, ""},
		{"finish", http.MethodPost, tripsPath + "/" + id.String() + "/finish", false, ""},
		{"health", http.MethodGet, "/health", false, ""},
		{"ready", http.MethodGet, "/ready", false, ""},
		{"invalid get UUID", http.MethodGet, tripsPath + "/bad", true, ""},
		{"invalid finish UUID", http.MethodPost, tripsPath + "/bad/finish", true, ""},
		{"invalid key", http.MethodPost, tripsPath, true, "bad"},
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
			writer := response.New()
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

func TestRouterRoutingErrors(t *testing.T) {
	router := httptransport.NewRouter(nil, nil, nil, nil, nil, response.New())
	for _, tc := range []struct {
		method, path, code, allow string
		status                    int
	}{
		{http.MethodGet, "/missing", "route_not_found", "", http.StatusNotFound},
		{http.MethodGet, tripsPath, "method_not_allowed", http.MethodPost, http.StatusMethodNotAllowed},
		{http.MethodPost, "/health", "method_not_allowed", http.MethodGet, http.StatusMethodNotAllowed},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			out := httptest.NewRecorder()
			router.ServeHTTP(out, httptest.NewRequest(tc.method, tc.path, nil))
			if out.Code != tc.status || out.Header().Get("Content-Type") != "application/problem+json" || out.Header().Get("Allow") != tc.allow {
				t.Fatalf("status %d, headers %v, body %s", out.Code, out.Header(), out.Body.String())
			}
			var body api.Problem
			if err := json.Unmarshal(out.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != tc.code || int(body.Status) != tc.status || body.Instance == nil || *body.Instance != tc.path || body.Type == "" || body.Title == "" {
				t.Fatalf("unexpected Problem: %s", out.Body.String())
			}
		})
	}
}
