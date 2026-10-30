package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric/noop"
)

func Tracing(next http.Handler) http.Handler {
	return otelhttp.NewHandler(
		next,
		"http.request",
		otelhttp.WithSpanNameFormatter(func(_ string, request *http.Request) string {
			return request.Method + " " + routeTemplate(request)
		}),
		otelhttp.WithMeterProvider(noop.NewMeterProvider()),
	)
}

func routeTemplate(request *http.Request) string {
	routeContext := chi.RouteContext(request.Context())
	if routeContext != nil {
		if pattern := routeContext.RoutePattern(); pattern != "" {
			return pattern
		}
	}

	return normalizedRoute(request.URL.Path)
}

func normalizedRoute(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "trips" {
		if _, err := uuid.Parse(parts[3]); err == nil {
			return "/api/v1/trips/{tripId}"
		}
	}
	if len(parts) == 5 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "trips" && parts[4] == "finish" {
		if _, err := uuid.Parse(parts[3]); err == nil {
			return "/api/v1/trips/{tripId}/finish"
		}
	}

	return path
}
