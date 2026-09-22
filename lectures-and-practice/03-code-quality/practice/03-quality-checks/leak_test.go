//go:build leaktest

package qualitychecks

import (
	"context"
	"testing"

	"go.uber.org/goleak"
)

func TestNoGoroutineLeaks(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(context.Background())
	Total(ctx, []int{1, 2, 3})
	cancel()
}
