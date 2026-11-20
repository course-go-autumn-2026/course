package task03

import (
	"context"
	"errors"
	"github.com/segmentio/kafka-go"
	"reflect"
	"testing"
)

func TestDLQ(t *testing.T) {
	for _, name := range []string{"success", "failure", "dlq-offline", "commit-error", "cancel", "deadline", "cancel-before"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if name == "cancel-before" {
				cancel()
			}
			var calls []string
			m := kafka.Message{Topic: "orders.task3", Partition: 2, Offset: 42, Key: []byte("o1"), Value: []byte("raw"), Headers: []kafka.Header{{Key: "trace", Value: []byte("t1")}}}
			sentinel := errors.New("offline")
			h := Hooks{Handle: func(context.Context, kafka.Message) error {
				calls = append(calls, "handle")
				if name == "success" || name == "commit-error" {
					return nil
				}
				if name == "cancel" {
					cancel()
					return ctx.Err()
				}
				if name == "deadline" {
					return context.DeadlineExceeded
				}
				return sentinel
			}, DeadLetter: func(_ context.Context, d DLQ) error {
				calls = append(calls, "dlq")
				if d.Topic != m.Topic || d.Partition != m.Partition || d.Offset != m.Offset || !reflect.DeepEqual(d.Key, m.Key) || !reflect.DeepEqual(d.Value, m.Value) || !reflect.DeepEqual(d.Headers, m.Headers) || d.Attempts != 1 || d.Reason != "offline" {
					t.Fatalf("envelope lost data: %+v", d)
				}
				if name == "dlq-offline" {
					return sentinel
				}
				return nil
			}, Commit: func(context.Context, kafka.Message) error {
				calls = append(calls, "commit")
				if name == "commit-error" {
					return sentinel
				}
				return nil
			}}
			err := Process(ctx, m, h)
			want := []string{"handle", "dlq", "commit"}
			switch name {
			case "success":
				want = []string{"handle", "commit"}
			case "dlq-offline":
				want = []string{"handle", "dlq"}
			case "commit-error":
				want = []string{"handle", "commit"}
			case "cancel", "deadline":
				want = []string{"handle"}
			case "cancel-before":
				want = nil
			}
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("calls %v want %v", calls, want)
			}
			switch name {
			case "success", "failure":
				if err != nil {
					t.Fatal(err)
				}
			case "cancel", "cancel-before":
				if !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
			case "deadline":
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
			default:
				if !errors.Is(err, sentinel) {
					t.Fatal(err)
				}
			}
		})
	}
}
