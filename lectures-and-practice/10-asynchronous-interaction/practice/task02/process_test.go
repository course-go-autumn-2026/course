package task02

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"reflect"
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		failures              int
		permanent, commitFail bool
		calls                 []string
	}{
		{"success", 0, false, false, []string{"handle", "commit"}},
		{"retry", 2, false, false, []string{"handle", "wait100ms", "handle", "wait200ms", "handle", "commit"}},
		{"exhausted", 9, false, false, []string{"handle", "wait100ms", "handle", "wait200ms", "handle"}},
		{"permanent", 9, true, false, []string{"handle"}},
		{"commit error", 0, false, true, []string{"handle", "commit"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			attempt := 0
			sentinel := errors.New("offline")
			h := Hooks{Handle: func(context.Context, kafka.Message) error {
				calls = append(calls, "handle")
				attempt++
				if attempt <= tc.failures {
					if tc.permanent {
						return fmt.Errorf("wrapped: %w", ErrPermanent)
					}
					return sentinel
				}
				return nil
			}, Wait: func(_ context.Context, d time.Duration) error { calls = append(calls, "wait"+d.String()); return nil }, Commit: func(context.Context, kafka.Message) error {
				calls = append(calls, "commit")
				if tc.commitFail {
					return sentinel
				}
				return nil
			}}
			err := Process(context.Background(), kafka.Message{}, h)
			switch {
			case tc.permanent:
				if !errors.Is(err, ErrPermanent) {
					t.Fatal(err)
				}
			case tc.failures >= 3 || tc.commitFail:
				if !errors.Is(err, sentinel) {
					t.Fatal(err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(calls, tc.calls) {
				t.Fatalf("got %v want %v", calls, tc.calls)
			}
		})
	}
}
func TestCancellation(t *testing.T) {
	for _, stage := range []string{"before", "handler", "wait", "deadline"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if stage == "before" {
				cancel()
			}
			h := Hooks{Handle: func(context.Context, kafka.Message) error {
				if stage == "handler" {
					cancel()
				}
				if stage == "deadline" {
					return context.DeadlineExceeded
				}
				return errors.New("temporary")
			}, Wait: func(context.Context, time.Duration) error { cancel(); return ctx.Err() }, Commit: func(context.Context, kafka.Message) error { t.Fatal("cancellation committed"); return nil }}
			want := error(context.Canceled)
			if stage == "deadline" {
				want = context.DeadlineExceeded
			}
			if !errors.Is(Process(ctx, kafka.Message{}, h), want) {
				t.Fatal("cancellation lost")
			}
		})
	}
}
