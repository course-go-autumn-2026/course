package container

import (
	"context"
	"strings"
	"testing"

	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/config"
)

func TestNewRejectsUnsupportedTelemetryProtocolBeforeDatabaseConnection(t *testing.T) {
	t.Parallel()

	_, err := New(context.Background(), config.Config{
		LogLevel: "info",
		Telemetry: config.TelemetryConfig{
			ServiceName: "trip-service",
			Endpoint:    "http://127.0.0.1:4318",
			Protocol:    "grpc",
		},
	})
	if err == nil {
		t.Fatal("New() error = nil")
	}
	if !strings.Contains(err.Error(), "OTEL_EXPORTER_OTLP_PROTOCOL") {
		t.Fatalf("error = %v", err)
	}
}
