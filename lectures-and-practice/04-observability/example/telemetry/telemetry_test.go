package telemetry

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	collectorlog "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectormetric "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
)

// One smoke check for the shared setup: all three signals survive shutdown,
// carry the per-process name, and the log retains its span's identifiers.
func TestSetupExportsCorrelatedSignals(t *testing.T) {
	var mu sync.Mutex
	traces := new(collectortrace.ExportTraceServiceRequest)
	metrics := new(collectormetric.ExportMetricsServiceRequest)
	logs := new(collectorlog.ExportLogsServiceRequest)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		var message proto.Message
		switch r.URL.Path {
		case "/v1/traces":
			message = traces
		case "/v1/metrics":
			message = metrics
		case "/v1/logs":
			message = logs
		default:
			t.Errorf("unexpected OTLP path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := (proto.UnmarshalOptions{Merge: true}).Unmarshal(body, message); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	t.Cleanup(receiver.Close)
	// Never inherit a per-signal endpoint or credentials from the caller's shell.
	for _, signal := range []string{"", "TRACES_", "METRICS_", "LOGS_"} {
		for _, option := range []string{"ENDPOINT", "HEADERS", "COMPRESSION"} {
			t.Setenv("OTEL_EXPORTER_OTLP_"+signal+option, "")
		}
	}
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", receiver.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_TRACES_SAMPLER", "always_on")
	t.Setenv("OTEL_SERVICE_NAME", "smoke-service")
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "60000")
	oldTracer, oldMeter, oldPropagator := otel.GetTracerProvider(), otel.GetMeterProvider(), otel.GetTextMapPropagator()
	defer func() {
		otel.SetTracerProvider(oldTracer)
		otel.SetMeterProvider(oldMeter)
		otel.SetTextMapPropagator(oldPropagator)
	}()

	logger, shutdown, err := Setup(context.Background(), "wrong-default-name")
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	t.Cleanup(func() {
		if closed {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			t.Error(err)
		}
	})
	ctx, span := otel.Tracer(scope).Start(context.Background(), "smoke")
	counter, err := otel.Meter(scope).Int64Counter("smoke.operations")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(ctx, 1)
	logger.InfoContext(ctx, "smoke completed")
	span.End()
	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	closed = true
	if err := shutdown(flushCtx); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(traces.ResourceSpans) == 0 || len(metrics.ResourceMetrics) == 0 || len(logs.ResourceLogs) == 0 {
		t.Fatalf("missing signal: traces=%d metrics=%d logs=%d", len(traces.ResourceSpans), len(metrics.ResourceMetrics), len(logs.ResourceLogs))
	}
	for _, res := range []*resourcepb.Resource{traces.ResourceSpans[0].Resource, metrics.ResourceMetrics[0].Resource, logs.ResourceLogs[0].Resource} {
		name := ""
		for _, attr := range res.Attributes {
			if attr.Key == "service.name" {
				name = attr.Value.GetStringValue()
			}
		}
		if name != "smoke-service" {
			t.Errorf("service.name = %q, want smoke-service", name)
		}
	}
	gotSpan := traces.ResourceSpans[0].ScopeSpans[0].Spans[0]
	gotLog := logs.ResourceLogs[0].ScopeLogs[0].LogRecords[0]
	if len(gotLog.TraceId) != 16 || !bytes.Equal(gotLog.TraceId, gotSpan.TraceId) || !bytes.Equal(gotLog.SpanId, gotSpan.SpanId) {
		t.Errorf("log is not correlated with exported span: log=%v span=%v", gotLog, gotSpan)
	}
	points := metrics.ResourceMetrics[0].ScopeMetrics[0].Metrics[0].GetSum().GetDataPoints()
	if len(points) != 1 || points[0].GetAsInt() != 1 {
		t.Errorf("counter must contain one operation, got %v", points)
	}
}
