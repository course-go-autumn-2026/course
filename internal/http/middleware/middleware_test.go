package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/DaleCoopTP/grishin_lab1/internal/http/middleware"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
)

const (
	tripRoute   = "/api/v1/trips/{tripId}"
	requestPath = "/api/v1/trips/specific-trip"
)

func TestRequestSignals(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		panics bool
		level  string
	}{
		{"success", http.StatusOK, false, "INFO"},
		{"client error", http.StatusNotFound, false, "INFO"},
		{"server error", http.StatusInternalServerError, false, "ERROR"},
		{"panic", http.StatusInternalServerError, true, "ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spans := tracetest.NewInMemoryExporter()
			provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(spans))
			reader := sdkmetric.NewManualReader()
			meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			t.Cleanup(func() {
				if err := provider.Shutdown(context.Background()); err != nil {
					t.Error(err)
				}
				if err := meter.Shutdown(context.Background()); err != nil {
					t.Error(err)
				}
			})
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			router := chi.NewRouter()
			router.Get(tripRoute, func(w http.ResponseWriter, r *http.Request) {
				assertMetrics(t, reader, r.Context(), 1, 0)
				if tc.panics {
					panic("private panic details")
				}
				w.WriteHeader(tc.status)
			})
			handler, err := middleware.Wrap(router, response.New(), logger, provider, meter.Meter("test"))
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodGet, requestPath, nil)
			req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
			out := httptest.NewRecorder()
			handler.ServeHTTP(out, req)
			if out.Code != tc.status {
				t.Fatalf("status = %d, want %d", out.Code, tc.status)
			}
			if tc.panics && (!strings.Contains(out.Body.String(), "internal_error") || strings.Contains(out.Body.String(), "private")) {
				t.Fatalf("unexpected panic response: %s", out.Body.String())
			}
			assertMetrics(t, reader, t.Context(), 0, tc.status)
			recorded := spans.GetSpans()
			if len(recorded) != 1 {
				t.Fatalf("HTTP spans: got %d, want 1", len(recorded))
			}
			span := recorded[0]
			if span.Name != "GET "+tripRoute || span.Parent.SpanID().String() != "00f067aa0ba902b7" {
				t.Fatalf("unexpected span name or parent: %q, %v", span.Name, span.Parent)
			}
			attrs := attribute.NewSet(span.Attributes...)
			assertLabel(t, attrs, "http.route", attribute.StringValue(tripRoute))
			assertLabel(t, attrs, "http.response.status_code", attribute.IntValue(tc.status))
			if (span.Status.Code == codes.Error) != (tc.status >= 500) {
				t.Fatalf("unexpected span status: %v", span.Status)
			}
			if tc.panics {
				assertLabel(t, attrs, "error.type", attribute.StringValue("panic"))
			}
			assertRequestLog(t, logs.String(), tc.status, tc.level, tc.panics)
		})
	}
}

func assertRequestLog(t *testing.T, text string, status int, level string, panicked bool) {
	t.Helper()
	if strings.Count(text, "\n") != 1 || strings.Contains(text, requestPath) || strings.Contains(text, "private panic details") {
		t.Fatalf("unexpected access log: %s", text)
	}
	var log struct {
		Level, Route, Method, Stack string
		Status                      int
	}
	if err := json.Unmarshal([]byte(text), &log); err != nil {
		t.Fatal(err)
	}
	if log.Level != level || log.Route != tripRoute || log.Method != http.MethodGet || log.Status != status || (log.Stack != "") != panicked {
		t.Fatalf("unexpected access log: %s", text)
	}
}

func assertMetrics(t *testing.T, reader *sdkmetric.ManualReader, ctx context.Context, active, status int) {
	t.Helper()
	var data metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &data); err != nil {
		t.Fatal(err)
	}
	seenActive, seenDuration := false, false
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			switch m.Name {
			case "http.server.active_requests":
				points := m.Data.(metricdata.Sum[int64]).DataPoints
				if len(points) != 1 || points[0].Value != int64(active) {
					t.Fatalf("active requests: %+v, want %d", points, active)
				}
				assertLabel(t, points[0].Attributes, "http.route", attribute.StringValue(tripRoute))
				seenActive = true
			case "http.server.request.duration":
				points := m.Data.(metricdata.Histogram[float64]).DataPoints
				if len(points) != 1 || points[0].Count != 1 || m.Unit != "s" {
					t.Fatalf("unexpected duration histogram: %+v", m)
				}
				assertLabel(t, points[0].Attributes, "http.route", attribute.StringValue(tripRoute))
				assertLabel(t, points[0].Attributes, "http.response.status_code", attribute.IntValue(status))
				seenDuration = true
			}
		}
	}
	if !seenActive || seenDuration != (status != 0) {
		t.Fatalf("missing or premature metrics: active %v, duration %v, status %d", seenActive, seenDuration, status)
	}
}

func assertLabel(t *testing.T, attrs attribute.Set, key attribute.Key, expected attribute.Value) {
	t.Helper()
	value, ok := attrs.Value(key)
	if !ok || value != expected {
		t.Fatalf("attribute %s: got %v, want %v", key, value, expected)
	}
}
