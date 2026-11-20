package task01

import (
	"context"
	"errors"
	"example.com/async-practice/internal/event"
	"github.com/segmentio/kafka-go"
)

type Writer interface {
	WriteMessages(context.Context, ...kafka.Message) error
}
type Reader interface {
	FetchMessage(context.Context) (kafka.Message, error)
	CommitMessages(context.Context, ...kafka.Message) error
}

// Publish validates and encodes e. Key is OrderID. Wait for WriteMessages result.
func Publish(ctx context.Context, w Writer, e event.OrderCreated) error {
	return errors.New("TODO task01 Publish")
}

// ConsumeOne: fetch -> handle -> commit. Do not commit on a handler error.
func ConsumeOne(ctx context.Context, r Reader, handle func(context.Context, kafka.Message) error) error {
	return errors.New("TODO task01 ConsumeOne")
}
