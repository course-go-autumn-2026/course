package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/config"
)

type Runtime struct {
	tracerProvider *trace.TracerProvider
	meterProvider  *metric.MeterProvider
}

func Setup(ctx context.Context, cfg config.TelemetryConfig, logger *slog.Logger) (*Runtime, error) {
	endpoint, insecure, err := endpoint(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	if cfg.Protocol != "http/protobuf" {
		return nil, fmt.Errorf("unsupported OTEL_EXPORTER_OTLP_PROTOCOL %q", cfg.Protocol)
	}

	traceOptions := []otlptracehttp.Option{otlptracehttp.WithEndpoint(endpoint)}
	metricOptions := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpoint(endpoint)}
	if insecure {
		traceOptions = append(traceOptions, otlptracehttp.WithInsecure())
		metricOptions = append(metricOptions, otlpmetrichttp.WithInsecure())
	}

	traceExporter, err := otlptracehttp.New(ctx, traceOptions...)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	metricExporter, err := otlpmetrichttp.New(ctx, metricOptions...)
	if err != nil {
		if shutdownErr := traceExporter.Shutdown(ctx); shutdownErr != nil {
			logger.Error("shutdown trace exporter", "error", shutdownErr)
		}
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}

	resource, err := newResource(cfg)
	if err != nil {
		return nil, err
	}
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(resource),
		trace.WithSampler(sampler(cfg.TracesSampler)),
	)
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return &Runtime{tracerProvider: tracerProvider, meterProvider: meterProvider}, nil
}

func (runtime *Runtime) Shutdown(ctx context.Context) error {
	return errors.Join(runtime.meterProvider.Shutdown(ctx), runtime.tracerProvider.Shutdown(ctx))
}

func endpoint(value string) (string, bool, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "", false, fmt.Errorf("parse OTEL_EXPORTER_OTLP_ENDPOINT: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", false, fmt.Errorf("unsupported OTEL_EXPORTER_OTLP_ENDPOINT scheme %q", parsed.Scheme)
	}

	return parsed.Host, parsed.Scheme == "http", nil
}

func newResource(cfg config.TelemetryConfig) (*resource.Resource, error) {
	attributes := []attribute.KeyValue{attribute.String("service.name", cfg.ServiceName)}
	for _, value := range strings.Split(cfg.ResourceAttributes, ",") {
		key, attributeValue, found := strings.Cut(value, "=")
		if !found || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("parse OTEL_RESOURCE_ATTRIBUTES value %q", value)
		}
		attributes = append(attributes, attribute.String(strings.TrimSpace(key), strings.TrimSpace(attributeValue)))
	}

	return resource.NewWithAttributes("", attributes...), nil
}

func sampler(value string) trace.Sampler {
	if value == "always_off" || value == "parentbased_always_off" {
		return trace.ParentBased(trace.NeverSample())
	}

	return trace.ParentBased(trace.AlwaysSample())
}
