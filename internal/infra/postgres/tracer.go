package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
)

type queryKey struct{}

type queryObservation struct {
	start                 time.Time
	operation, collection string
}

type queryTracer struct {
	traces   *otelpgx.Tracer
	duration metric.Float64Histogram
}

func NewQueryTracer(provider trace.TracerProvider, meter metric.Meter) (pgx.QueryTracer, error) {
	duration, err := meter.Float64Histogram("db.client.operation.duration", metric.WithUnit("s"))
	if err != nil {
		return nil, fmt.Errorf("create database duration histogram: %w", err)
	}
	tracer := otelpgx.NewTracer(
		otelpgx.WithTracerProvider(provider),
		otelpgx.WithMeterProvider(noop.NewMeterProvider()),
		otelpgx.WithDisableSQLStatementInAttributes(),
		otelpgx.WithDisableConnectionDetailsInAttributes(),
	)
	return &queryTracer{traces: tracer, duration: duration}, nil
}

func (t *queryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	observation := queryObservation{
		start:      time.Now(),
		operation:  otelpgx.SQLOperationName(ctx, data.SQL),
		collection: queryCollection(data.SQL),
	}
	ctx = t.traces.TraceQueryStart(ctx, conn, data)
	trace.SpanFromContext(ctx).SetAttributes(observation.attributes()...)
	return context.WithValue(ctx, queryKey{}, observation)
}

func (t *queryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	observation := ctx.Value(queryKey{}).(queryObservation)
	attrs := observation.attributes()
	if data.Err != nil {
		errorType := queryErrorType(data.Err)
		attr := attribute.String("error.type", errorType)
		attrs = append(attrs, attr)
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(attr)
		span.SetStatus(codes.Error, errorType)
		data.Err = errors.New(errorType)
	}
	t.duration.Record(ctx, time.Since(observation.start).Seconds(), metric.WithAttributes(attrs...))
	t.traces.TraceQueryEnd(ctx, conn, data)
}

func (o queryObservation) attributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("db.operation.name", o.operation),
		attribute.String("db.collection.name", o.collection),
	}
}

func queryCollection(query string) string {
	fields := strings.Fields(query)
	for i := 1; i < len(fields); i++ {
		switch strings.ToUpper(fields[i-1]) {
		case "FROM", "INTO", "UPDATE":
			name := strings.ToLower(fields[i])
			switch name {
			case "trips", "trip_status_history", "trip_idempotency_keys":
				return name
			}
		}
	}
	return "none"
}

func queryErrorType(err error) string {
	var pgErr *pgconn.PgError
	switch {
	case errors.As(err, &pgErr):
		return pgErr.Code
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	default:
		return "unknown"
	}
}
