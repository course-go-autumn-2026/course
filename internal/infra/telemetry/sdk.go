package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const exportTimeout = 2 * time.Second

type SDK struct {
	Traces  *sdktrace.TracerProvider
	Metrics *sdkmetric.MeterProvider
}

func NewSDK(ctx context.Context, logger *slog.Logger) *SDK {
	attributes := []attribute.KeyValue{
		attribute.String("service.name", serviceName),
		attribute.String("service.namespace", "tripgo"),
		attribute.String("deployment.environment", "local"),
		attribute.String("service.version", buildVersion()),
	}
	res, err := resource.New(ctx, resource.WithAttributes(attributes...), resource.WithFromEnv())
	if err != nil {
		logger.WarnContext(ctx, "OTel resource configuration failed", "error.type", fmt.Sprintf("%T", err))
		res = resource.NewSchemaless(attributes...)
	}
	traceOpts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	metricOpts := []sdkmetric.Option{sdkmetric.WithResource(res)}
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		traces, metrics, err := newExporters(ctx, os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"))
		if err != nil {
			logger.WarnContext(ctx, "OTLP exporter configuration failed", "error.type", fmt.Sprintf("%T", err))
		}
		if traces != nil {
			traceOpts = append(traceOpts, sdktrace.WithBatcher(traceExporter{traces, logger}, sdktrace.WithExportTimeout(exportTimeout)))
		}
		if metrics != nil {
			metricOpts = append(metricOpts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter{metrics, logger}, sdkmetric.WithTimeout(exportTimeout))))
		}
	}
	return &SDK{
		Traces:  sdktrace.NewTracerProvider(traceOpts...),
		Metrics: sdkmetric.NewMeterProvider(metricOpts...),
	}
}

func (s *SDK) Shutdown(ctx context.Context) error {
	return errors.Join(s.Metrics.Shutdown(ctx), s.Traces.Shutdown(ctx))
}

func newExporters(ctx context.Context, protocol string) (sdktrace.SpanExporter, sdkmetric.Exporter, error) {
	var traces sdktrace.SpanExporter
	var metrics sdkmetric.Exporter
	var traceErr, metricErr error
	switch protocol {
	case "", "http/protobuf":
		traces, traceErr = otlptracehttp.New(ctx)
		metrics, metricErr = otlpmetrichttp.New(ctx)
	case "grpc":
		traces, traceErr = otlptracegrpc.New(ctx)
		metrics, metricErr = otlpmetricgrpc.New(ctx)
	default:
		return nil, nil, fmt.Errorf("unsupported OTLP protocol %q", protocol)
	}
	if traceErr != nil {
		traces = nil
	}
	if metricErr != nil {
		metrics = nil
	}
	return traces, metrics, errors.Join(traceErr, metricErr)
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return info.Main.Version
}

type traceExporter struct {
	sdktrace.SpanExporter
	logger *slog.Logger
}

func (e traceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if err := e.SpanExporter.ExportSpans(ctx, spans); err != nil {
		e.logger.WarnContext(ctx, "OTLP trace export failed", "error.type", fmt.Sprintf("%T", err))
	}
	return nil
}

type metricExporter struct {
	sdkmetric.Exporter
	logger *slog.Logger
}

func (e metricExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	if err := e.Exporter.Export(ctx, data); err != nil {
		e.logger.WarnContext(ctx, "OTLP metric export failed", "error.type", fmt.Sprintf("%T", err))
	}
	return nil
}
