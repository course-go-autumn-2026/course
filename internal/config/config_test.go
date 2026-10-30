package config

import "testing"

func TestLoadReadsTelemetryConfiguration(t *testing.T) {
	setRequiredEnvironment(t)
	t.Setenv("OTEL_SERVICE_NAME", "trip-service")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.namespace=tripgo,deployment.environment=local")
	t.Setenv("OTEL_TRACES_SAMPLER", "parentbased_always_on")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Telemetry.ServiceName != "trip-service" {
		t.Errorf("Telemetry.ServiceName = %q, want %q", cfg.Telemetry.ServiceName, "trip-service")
	}
	if cfg.Telemetry.Endpoint != "http://localhost:4318" {
		t.Errorf("Telemetry.Endpoint = %q, want %q", cfg.Telemetry.Endpoint, "http://localhost:4318")
	}
	if cfg.Telemetry.Protocol != "http/protobuf" {
		t.Errorf("Telemetry.Protocol = %q, want %q", cfg.Telemetry.Protocol, "http/protobuf")
	}
	if cfg.Telemetry.ResourceAttributes != "service.namespace=tripgo,deployment.environment=local" {
		t.Errorf("Telemetry.ResourceAttributes = %q, want resource attributes", cfg.Telemetry.ResourceAttributes)
	}
	if cfg.Telemetry.TracesSampler != "parentbased_always_on" {
		t.Errorf("Telemetry.TracesSampler = %q, want %q", cfg.Telemetry.TracesSampler, "parentbased_always_on")
	}
}

func TestLoadUsesDefaultHTTPTimeouts(t *testing.T) {
	setEnvironmentWithoutHTTPTimeouts(t)
	t.Setenv("OTEL_SERVICE_NAME", "trip-service")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.namespace=tripgo,deployment.environment=local")
	t.Setenv("OTEL_TRACES_SAMPLER", "parentbased_always_on")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPReadTimeout.String() != "5s" {
		t.Errorf("HTTPReadTimeout = %s, want 5s", cfg.HTTPReadTimeout)
	}
	if cfg.HTTPReadHeaderTimeout.String() != "2s" {
		t.Errorf("HTTPReadHeaderTimeout = %s, want 2s", cfg.HTTPReadHeaderTimeout)
	}
	if cfg.HTTPWriteTimeout.String() != "10s" {
		t.Errorf("HTTPWriteTimeout = %s, want 10s", cfg.HTTPWriteTimeout)
	}
	if cfg.HTTPIdleTimeout.String() != "1m0s" {
		t.Errorf("HTTPIdleTimeout = %s, want 1m0s", cfg.HTTPIdleTimeout)
	}
}

func setRequiredEnvironment(t *testing.T) {
	t.Helper()
	setEnvironmentWithoutHTTPTimeouts(t)
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "2s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "10s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "60s")
}

func setEnvironmentWithoutHTTPTimeouts(t *testing.T) {
	t.Helper()
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("DATABASE_URL", "postgres://tripgo:tripgo@localhost:5432/tripgo?sslmode=disable")
	t.Setenv("DATABASE_MAX_CONNS", "10")
	t.Setenv("DATABASE_MIN_CONNS", "2")
	t.Setenv("DATABASE_MAX_CONN_LIFETIME", "30m")
	t.Setenv("DATABASE_CONNECT_TIMEOUT", "5s")
	t.Setenv("DATABASE_QUERY_TIMEOUT", "3s")
}
