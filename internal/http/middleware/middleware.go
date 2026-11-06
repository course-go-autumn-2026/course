package middleware

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
)

const (
	methodKey = "http.request.method"
	routeKey  = "http.route"
	statusKey = "http.response.status_code"
)

type stateKey struct{}

type requestState struct {
	method, route string
	status        int
	log           func(int, time.Duration, []byte)
	record        func(int, time.Duration)
}

func Wrap(router chi.Router, writer *response.Writer, logger *slog.Logger, tracer trace.TracerProvider, meter metric.Meter) (http.Handler, error) {
	duration, err := meter.Float64Histogram("http.server.request.duration", metric.WithUnit("s"))
	if err != nil {
		return nil, fmt.Errorf("create HTTP duration histogram: %w", err)
	}
	active, err := meter.Int64UpDownCounter("http.server.active_requests", metric.WithUnit("{request}"))
	if err != nil {
		return nil, fmt.Errorf("create active requests counter: %w", err)
	}
	measured := metrics(router, duration, active)
	logged := logging(measured, logger)
	traced := tracing(logged, tracer)
	return recoverRequests(router, traced, writer, logger), nil
}

func recoverRequests(router chi.Router, next http.Handler, writer *response.Writer, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		state := &requestState{method: requestMethod(r.Method), route: routePattern(router, r)}
		r = r.WithContext(context.WithValue(r.Context(), stateKey{}, state))
		w = captureStatus(w, state)
		defer func() {
			var stack []byte
			if recover() != nil {
				stack = debug.Stack()
				if state.status == 0 {
					writer.Error(w, r, errors.New("HTTP handler panic"))
				}
			}
			status := state.responseStatus()
			elapsed := time.Since(start)
			if state.record != nil {
				state.record(status, elapsed)
			}
			if state.log != nil {
				state.log(status, elapsed, stack)
			} else {
				logRequest(r.Context(), logger, state, status, elapsed, stack)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func tracing(next http.Handler, provider trace.TracerProvider) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.Context().Value(stateKey{}).(*requestState)
		span := trace.SpanFromContext(r.Context())
		span.SetAttributes(attribute.String(routeKey, state.route))
		finished := false
		defer func() {
			status := state.responseStatus()
			if !finished && state.status == 0 {
				status = http.StatusInternalServerError
			}
			if status >= http.StatusInternalServerError || !finished {
				span.SetStatus(codes.Error, "HTTP server error")
				errorType := strconv.Itoa(status)
				if !finished {
					errorType = "panic"
				}
				span.SetAttributes(attribute.Int(statusKey, status), attribute.String("error.type", errorType))
			}
		}()
		next.ServeHTTP(w, r)
		finished = true
	})
	return otelhttp.NewHandler(handler, "HTTP request",
		otelhttp.WithTracerProvider(provider),
		otelhttp.WithMeterProvider(noop.NewMeterProvider()),
		otelhttp.WithPropagators(propagation.TraceContext{}),
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			state := r.Context().Value(stateKey{}).(*requestState)
			return state.method + " " + state.route
		}),
	)
}

func logging(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		state := ctx.Value(stateKey{}).(*requestState)
		state.log = func(status int, elapsed time.Duration, stack []byte) {
			logRequest(ctx, logger, state, status, elapsed, stack)
		}
		next.ServeHTTP(w, r)
	})
}

func metrics(next http.Handler, duration metric.Float64Histogram, active metric.Int64UpDownCounter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		state := ctx.Value(stateKey{}).(*requestState)
		method := attribute.String(methodKey, state.method)
		route := attribute.String(routeKey, state.route)
		activeAttrs := metric.WithAttributes(method, route)
		active.Add(ctx, 1, activeAttrs)
		state.record = func(status int, elapsed time.Duration) {
			active.Add(ctx, -1, activeAttrs)
			duration.Record(ctx, elapsed.Seconds(), metric.WithAttributes(method, route, attribute.Int(statusKey, status)))
		}
		next.ServeHTTP(w, r)
	})
}

func logRequest(ctx context.Context, logger *slog.Logger, state *requestState, status int, elapsed time.Duration, stack []byte) {
	level := slog.LevelInfo
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
	}
	attrs := []slog.Attr{
		slog.String("method", state.method),
		slog.String("route", state.route),
		slog.Int("status", status),
		slog.Int64("duration_ms", elapsed.Milliseconds()),
	}
	if len(stack) > 0 {
		attrs = append(attrs, slog.String("error.type", "panic"), slog.String("stack", string(stack)))
	}
	logger.LogAttrs(ctx, level, "request handled", attrs...)
}

func routePattern(router chi.Router, r *http.Request) string {
	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}
	ctx := chi.NewRouteContext()
	if router.Match(ctx, r.Method, path) {
		return ctx.RoutePattern()
	}
	return "unmatched"
}

func requestMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return method
	default:
		return "_OTHER"
	}
}

func (s *requestState) responseStatus() int {
	if s.status == 0 {
		return http.StatusOK
	}
	return s.status
}

func captureStatus(w http.ResponseWriter, state *requestState) http.ResponseWriter {
	commit := func(status int) {
		if state.status == 0 {
			state.status = status
		}
	}
	return httpsnoop.Wrap(w, httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(status int) {
				next(status)
				if status >= 200 || status == http.StatusSwitchingProtocols {
					commit(status)
				}
			}
		},
		Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return func(data []byte) (int, error) {
				commit(http.StatusOK)
				return next(data)
			}
		},
		Flush: func(next httpsnoop.FlushFunc) httpsnoop.FlushFunc {
			return func() {
				commit(http.StatusOK)
				next()
			}
		},
		ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
			return func(r io.Reader) (int64, error) {
				commit(http.StatusOK)
				return next(r)
			}
		},
	})
}
