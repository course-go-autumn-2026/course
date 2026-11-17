package packing

import (
	"context"
	"sync"
	"time"
)

// Задание 1: общая map и Mutex. Каркас ожидания уже дан.
func PackAll(items []PackRequest, work Packer) []PackedOrder {
	if len(items) <= 0 {
		return nil
	}

	results := make(map[int]PackedOrder)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for index, item := range items {
		wg.Add(1)
		go func(index int, item PackRequest) {
			defer wg.Done()
			res := work(item)
			mu.Lock()
			defer mu.Unlock()
			results[index] = res
		}(index, item)
	}
	wg.Wait()
	output := make([]PackedOrder, len(items))
	for index := range items {
		output[index] = results[index]
	}
	return output
}

// Задание 2: Fan-out / Fan-in. Каркас объединения потоков уже дан.
func FanOut(in <-chan PackRequest, workers int, work Packer) []<-chan PackedOrder {
	if workers <= 0 {
		return nil
	}
	result := make([]<-chan PackedOrder, workers)
	for i := 0; i < workers; i++ {
		ch := make(chan PackedOrder)
		result[i] = ch
		go func() {
			defer close(ch)
			for req := range in {
				ans := work(req)
				ch <- ans
			}
		}()
	}

	return result
}
func FanIn(streams ...<-chan PackedOrder) <-chan PackedOrder {
	out := make(chan PackedOrder)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan PackedOrder) {
			defer wg.Done()
			for v := range stream {
				out <- v
			}
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
func RunPackingReport(ctx context.Context, interval time.Duration, report func()) error {
	if interval <= 0 {
		return ErrInvalidInterval
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			report()
		}
	}
}
