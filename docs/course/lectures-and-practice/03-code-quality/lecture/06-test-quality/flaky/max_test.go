package flakydemo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxValueKey(t *testing.T) {
	if os.Getenv("RUN_FLAKY_DEMO") != "1" {
		t.Skip("set RUN_FLAKY_DEMO=1 and use -count to expose the flaky assertion")
	}

	values := map[string]int{
		"apple":  3,
		"banana": 5,
		"cherry": 5,
	}

	assert.Equal(t, "banana", MaxValueKey(values))
}

func TestMaxValueKey_UniqueMaximum(t *testing.T) {
	values := map[string]int{
		"apple":  3,
		"banana": 8,
		"cherry": 5,
	}

	assert.Equal(t, "banana", MaxValueKey(values))
}
