package replay

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/async-practice/task03"
	"github.com/segmentio/kafka-go"
	"reflect"
	"testing"
)

type writer struct {
	messages []kafka.Message
	err      error
}

func (w *writer) WriteMessages(_ context.Context, m ...kafka.Message) error {
	w.messages = append(w.messages, m...)
	return w.err
}
func TestReplay(t *testing.T) {
	d := task03.DLQ{Topic: "orders.task3", Partition: 1, Offset: 5, Key: []byte("o1"), Value: []byte(`{"id":"e1"}`), Headers: []kafka.Header{{Key: "trace", Value: []byte("t1")}}, Attempts: 1, Reason: "rejected"}
	b, _ := json.Marshal(d)
	w := &writer{}
	if err := Publish(context.Background(), w, b); err != nil {
		t.Fatal(err)
	}
	m := w.messages[0]
	if !reflect.DeepEqual(m.Key, d.Key) || !reflect.DeepEqual(m.Value, d.Value) || !reflect.DeepEqual(m.Headers, d.Headers) {
		t.Fatal("changed original")
	}
	sentinel := errors.New("offline")
	w.err = sentinel
	if !errors.Is(Publish(context.Background(), w, b), sentinel) {
		t.Fatal("write error lost")
	}
	for _, topic := range []string{"orders.dlq", "other", ""} {
		d.Topic = topic
		b, _ = json.Marshal(d)
		before := len(w.messages)
		if Publish(context.Background(), w, b) == nil || len(w.messages) != before {
			t.Fatal("unsafe source topic accepted")
		}
	}
	if Publish(context.Background(), w, []byte("bad")) == nil {
		t.Fatal("invalid JSON accepted")
	}
}
