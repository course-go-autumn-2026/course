package task01

import (
	"context"
	"encoding/json"
	"errors"
	"example.com/async-practice/internal/event"
	"github.com/segmentio/kafka-go"
	"testing"
)

type fake struct {
	m                             kafka.Message
	writeErr, fetchErr, commitErr error
	calls                         []string
}

func (f *fake) WriteMessages(c context.Context, m ...kafka.Message) error {
	f.calls = append(f.calls, "write")
	f.m = m[0]
	return f.writeErr
}
func (f *fake) FetchMessage(c context.Context) (kafka.Message, error) {
	f.calls = append(f.calls, "fetch")
	return f.m, f.fetchErr
}
func (f *fake) CommitMessages(c context.Context, m ...kafka.Message) error {
	f.calls = append(f.calls, "commit")
	return f.commitErr
}
func TestPublish(t *testing.T) {
	ctx := context.Background()
	e := event.OrderCreated{ID: "e1", OrderID: "o1", Address: "Центр", Version: 1}
	f := &fake{}
	if err := Publish(ctx, f, e); err != nil {
		t.Fatal(err)
	}
	var got event.OrderCreated
	if err := json.Unmarshal(f.m.Value, &got); err != nil || got != e || string(f.m.Key) != "o1" {
		t.Fatalf("bad envelope: %+v", f.m)
	}
	sentinel := errors.New("offline")
	f.writeErr = sentinel
	if !errors.Is(Publish(ctx, f, e), sentinel) {
		t.Fatal("publish error lost")
	}
	f.calls = nil
	e.ID = ""
	if Publish(ctx, f, e) == nil || len(f.calls) > 0 {
		t.Fatal("invalid event published")
	}
}
func TestCommitOrder(t *testing.T) {
	for _, fail := range []string{"", "fetch", "handle", "commit"} {
		t.Run(fail, func(t *testing.T) {
			sentinel := errors.New(fail)
			f := &fake{}
			if fail == "fetch" {
				f.fetchErr = sentinel
			}
			if fail == "commit" {
				f.commitErr = sentinel
			}
			err := ConsumeOne(context.Background(), f, func(context.Context, kafka.Message) error {
				f.calls = append(f.calls, "handle")
				if fail == "handle" {
					return sentinel
				}
				return nil
			})
			if fail == "" && err != nil {
				t.Fatal(err)
			}
			if fail != "" && !errors.Is(err, sentinel) {
				t.Fatal("error not propagated")
			}
			want := []string{"fetch", "handle", "commit"}
			if fail == "fetch" {
				want = want[:1]
			}
			if fail == "handle" {
				want = want[:2]
			}
			if len(f.calls) != len(want) {
				t.Fatalf("order %v", f.calls)
			}
			for i := range want {
				if f.calls[i] != want[i] {
					t.Fatalf("order %v", f.calls)
				}
			}
		})
	}
}
