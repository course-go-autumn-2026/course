package qualitychecks

import (
	"context"
	"testing"
)

func TestTotal(t *testing.T) {
	values := make([]int, 1000)
	for index := range values {
		values[index] = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	got := Total(ctx, values)
	cancel()

	if got != len(values) {
		t.Fatalf("Total() = %d, want %d", got, len(values))
	}
}
