package replay

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/async-practice/task03"
	"github.com/segmentio/kafka-go"
)

type Writer interface {
	WriteMessages(context.Context, ...kafka.Message) error
}

// Publish is supplied tooling: only the task03 source topic is allowed.
func Publish(ctx context.Context, w Writer, raw []byte) error {
	var d task03.DLQ
	if err := json.Unmarshal(raw, &d); err != nil {
		return err
	}
	if d.Topic != "orders.task3" || d.Partition < 0 || d.Offset < 0 || len(d.Value) == 0 || d.Attempts < 1 || d.Reason == "" {
		return errors.New("invalid DLQ envelope or disallowed source topic")
	}
	return w.WriteMessages(ctx, kafka.Message{Key: d.Key, Value: d.Value, Headers: d.Headers})
}
