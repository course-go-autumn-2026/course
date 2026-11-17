package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type metricsSpy struct {
	startedRoute  string
	finishedRoute string
	status        int
}

func (spy *metricsSpy) RequestStarted(_ context.Context, _, route string) {
	spy.startedRoute = route
}

func (spy *metricsSpy) RequestFinished(_ context.Context, _, route string, status int, _ float64) {
	spy.finishedRoute = route
	spy.status = status
}

func TestMetricsRecordsResponseStatusAndRouteTemplate(t *testing.T) {
	t.Parallel()

	spy := &metricsSpy{}
	router := chi.NewRouter()
	router.Use(Metrics(spy))
	router.Get("/api/v1/trips/{tripId}", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusCreated)
	})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/trips/11111111-1111-4111-8111-111111111111", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if spy.finishedRoute != "/api/v1/trips/{tripId}" {
		t.Fatalf("route = %q", spy.finishedRoute)
	}
	if spy.startedRoute != "/api/v1/trips/{tripId}" {
		t.Fatalf("started route = %q", spy.startedRoute)
	}
	if spy.status != http.StatusCreated {
		t.Fatalf("recorded status = %d", spy.status)
	}
}

func TestRecoverConvertsPanicToProblem(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
	handler := Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("unexpected")
	}))
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if response.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
	}
}
