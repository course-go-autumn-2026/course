// Пакет racedemo содержит небезопасный и защищённый счётчики для демонстрации детектора гонок.
package racedemo

import "sync"

type Counter struct {
	Value int
}

func (c *Counter) Increment() {
	c.Value++
}

type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}
