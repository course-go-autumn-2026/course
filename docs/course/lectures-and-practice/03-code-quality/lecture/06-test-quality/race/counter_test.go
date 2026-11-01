package racedemo

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	workers    = 20
	increments = 100
)

func TestCounterRace(t *testing.T) {
	counter := &Counter{}
	runConcurrently(counter.Increment)
	t.Logf("without -race this value is not reliable: %d", counter.Value)
}

func TestSafeCounter(t *testing.T) {
	counter := &SafeCounter{}
	runConcurrently(counter.Increment)
	assert.Equal(t, workers*increments, counter.Value())
}

func runConcurrently(increment func()) {
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for range increments {
				increment()
			}
		}()
	}
	group.Wait()
}
