// Package telemetry configures OpenTelemetry once per service process.
package telemetry

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

const scope = "example.com/food-observability"

// Setup installs process-wide providers and propagation, and returns a context-aware
// logger. The caller must drain HTTP handlers before calling shutdown with a fresh,
// bounded context. OTEL_SERVICE_NAME overrides defaultName; other settings use OTEL_*.
func Setup(ctx context.Context, defaultName string) (_ *slog.Logger, shutdown func(context.Context) error, err error) {
	var closers []func(context.Context) error
	shutdown = func(ctx context.Context) error {
		var result error
		// Logs first: an unavailable Collector must not delay stdout export.
		for i := len(closers) - 1; i >= 0; i-- {
			result = errors.Join(result, closers[i](ctx))
		}
		return result
	}
	defer func() {
		if err != nil {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err = errors.Join(err, shutdown(cleanupCtx))
		}
	}()
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return nil, shutdown, errors.New("set OTEL_EXPORTER_OTLP_ENDPOINT: load the generated .env first")
	}
	if protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"); protocol != "" && protocol != "http/protobuf" {
		return nil, shutdown, errors.New("this example uses OTLP/HTTP protobuf exporters, not gRPC")
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(defaultName)),
		resource.WithFromEnv(), // Explicit per-process OTEL_SERVICE_NAME wins.
		resource.WithTelemetrySDK(),
		resource.WithAttributes(semconv.ServiceInstanceID(rand.Text())),
	)
	if err != nil {
		return nil, shutdown, err
	}
	traceExporter, err := otlptracehttp.New(ctx, otlptracehttp.WithTimeout(2*time.Second))
	if err != nil {
		return nil, shutdown, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(time.Second)),
	)
	closers = append(closers, tp.Shutdown)

	metricExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithTimeout(2*time.Second))
	if err != nil {
		return nil, shutdown, err
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		// OTEL_METRIC_EXPORT_INTERVAL is read by the SDK, in milliseconds.
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
	)
	closers = append(closers, mp.Shutdown)

	consoleExporter, err := stdoutlog.New()
	if err != nil {
		return nil, shutdown, err
	}
	logExporter, err := otlploghttp.New(ctx, otlploghttp.WithTimeout(2*time.Second))
	if err != nil {
		return nil, shutdown, errors.Join(err, consoleExporter.Shutdown(ctx))
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(consoleExporter)),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)
	closers = append(closers, lp.Shutdown)
	logger := otelslog.NewLogger(scope, otelslog.WithLoggerProvider(lp))
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	// Context travels with the business HTTP request, not over OTLP.
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return logger, shutdown, nil
}
