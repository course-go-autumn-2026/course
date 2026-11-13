package telemetry

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type traceHandler struct {
	next    slog.Handler
	service string
}

func NewLogger(writer io.Writer, level slog.Level, service string) *slog.Logger {
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	return slog.New(&traceHandler{next: handler, service: service})
}

func (handler *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.next.Enabled(ctx, level)
}

func (handler *traceHandler) Handle(ctx context.Context, record slog.Record) error {
	record.AddAttrs(slog.String("service", handler.service))
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
		)
	}

	return handler.next.Handle(ctx, record)
}

func (handler *traceHandler) WithAttrs(attributes []slog.Attr) slog.Handler {
	return &traceHandler{next: handler.next.WithAttrs(attributes), service: handler.service}
}

func (handler *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{next: handler.next.WithGroup(name), service: handler.service}
}
