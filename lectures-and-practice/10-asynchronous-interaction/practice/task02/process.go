package task02

import (
	"context"
	"errors"
	"example.com/async-practice/internal/processing"
	"github.com/segmentio/kafka-go"
	"time"
)

var ErrPermanent = processing.ErrPermanent

type Hooks struct {
	Handle func(context.Context, kafka.Message) error
	Commit func(context.Context, kafka.Message) error
	Wait   func(context.Context, time.Duration) error
}

// Process: 3 attempts total, waits 100ms/200ms. Commit only after success.
// Permanent errors, exhaustion and cancellation return without commit; no DLQ here.
func Process(ctx context.Context, m kafka.Message, h Hooks) error { return errors.New("TODO task02") }
