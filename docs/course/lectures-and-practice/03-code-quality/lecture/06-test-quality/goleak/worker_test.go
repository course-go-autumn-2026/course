package goleakdemo

import (
	"context"
	"os"
	"testing"

	"go.uber.org/goleak"
)

func TestWorkerStops(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	done := StartWorker(ctx)
	cancel()
	<-done
}

func TestLeakyWorker(t *testing.T) {
	if os.Getenv("RUN_LEAK_DEMO") != "1" {
		t.Skip("set RUN_LEAK_DEMO=1 to run the intentionally failing demo")
	}
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	StartLeakyWorker()
}
