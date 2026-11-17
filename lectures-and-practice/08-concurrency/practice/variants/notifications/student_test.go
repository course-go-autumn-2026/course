package notifications

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// Задание 1
func TestSendAll(t *testing.T) {
	items := make([]Notification, 100)
	var calls atomic.Int32

	work := func(n Notification) DeliveryResult {
		calls.Add(1)
		return DeliveryResult{}
	}

	results := SendAll(items, work)

	if len(results) != len(items) {
		t.Fatalf("expected %d results, got %d", len(items), len(results))
	}

	if calls.Load() != int32(len(items)) {
		t.Fatalf("expected %d calls, got %d", len(items), calls.Load())
	}
}

// Задание 2
func TestFanOutFanIn(t *testing.T) {
	const count = 100
	const workers = 4

	in := make(chan Notification, count)

	for i := 0; i < count; i++ {
		in <- Notification{}
	}
	close(in)

	var calls atomic.Int32

	work := func(n Notification) DeliveryResult {
		calls.Add(1)
		return DeliveryResult{}
	}

	streams := FanOut(in, workers, work)

	if len(streams) != workers {
		t.Fatalf("expected %d streams, got %d", workers, len(streams))
	}

	out := FanIn(streams...)

	received := 0
	for range out {
		received++
	}

	if received != count {
		t.Errorf("expected %d results, got %d", count, received)
	}

	if calls.Load() != count {
		t.Errorf("expected %d calls, got %d", count, calls.Load())
	}
}

func TestFanIn(t *testing.T) {
	ch1 := make(chan DeliveryResult, 2)
	ch2 := make(chan DeliveryResult, 1)

	ch1 <- DeliveryResult{}
	ch1 <- DeliveryResult{}
	ch2 <- DeliveryResult{}

	close(ch1)
	close(ch2)

	out := FanIn(ch1, ch2)

	count := 0
	for range out {
		count++
	}

	if count != 3 {
		t.Fatalf("expected 3 results, got %d", count)
	}
}

func TestFanInEmpty(t *testing.T) {
	out := FanIn()

	select {
	case _, ok := <-out:
		if ok {
			t.Fatal("expected closed channel")
		}
	case <-time.After(time.Second):
		t.Fatal("channel was not closed")
	}
}

// Задание 3
func TestRunNotificationReportInvalidInterval(t *testing.T) {
	ctx := context.Background()

	err := RunNotificationReport(ctx, 0, func() {})

	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestRunNotificationReportCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false

	err := RunNotificationReport(ctx, time.Second, func() {
		called = true
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if called {
		t.Fatal("report should not be called")
	}
}

func TestRunNotificationReportPeriodic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	calls := 0

	err := RunNotificationReport(ctx, 10*time.Millisecond, func() {
		calls++
		if calls == 3 {
			cancel()
		}
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if calls < 3 {
		t.Fatalf("expected at least 3 calls, got %d", calls)
	}
}
