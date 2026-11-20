package task03

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
)

type DLQ struct {
	Topic     string         `json:"topic"`
	Partition int            `json:"partition"`
	Offset    int64          `json:"offset"`
	Key       []byte         `json:"key"`
	Value     []byte         `json:"value"`
	Headers   []kafka.Header `json:"headers"`
	Attempts  int            `json:"attempts"`
	Reason    string         `json:"reason"`
}
type Hooks struct {
	Handle     func(context.Context, kafka.Message) error
	DeadLetter func(context.Context, DLQ) error
	Commit     func(context.Context, kafka.Message) error
}

// Process makes ONE handler attempt for this isolated exercise.
// Commit after success or acknowledged DLQ only. Never route cancellation to DLQ.
func Process(ctx context.Context, m kafka.Message, h Hooks) error { return errors.New("TODO task03") }
