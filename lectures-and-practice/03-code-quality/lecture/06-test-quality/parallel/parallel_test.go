package paralleldemo

import (
	"testing"
	"time"
)

func TestSlowCalculationA(t *testing.T) {
	t.Parallel()
	time.Sleep(250 * time.Millisecond)
}

func TestSlowCalculationB(t *testing.T) {
	t.Parallel()
	time.Sleep(250 * time.Millisecond)
}
