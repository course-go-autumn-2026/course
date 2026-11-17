package stock

import (
	"context"
	"sync"
	"time"
)

// Задание 1: общая map и Mutex. Каркас ожидания уже дан.
func CheckAllIngredients(items []Ingredient, work StockChecker) []StockStatus {
	results := make(map[int]StockStatus)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for index, item := range items {
		wg.Add(1)
		go func(index int, item Ingredient) {
			defer wg.Done()
			// TODO: вызовите work вне блокировки и сохраните результат под мьютексом.
			status := work(item)

			mu.Lock()
			defer mu.Unlock()
			results[index] = status
		}(index, item)
	}
	wg.Wait()
	output := make([]StockStatus, len(items))
	for index := range items {
		output[index] = results[index]
	}
	return output
}

// Задание 2: Fan-out / Fan-in. Каркас объединения потоков уже дан.
func FanOut(in <-chan Ingredient, workers int, work StockChecker) []<-chan StockStatus {
	// TODO: workers горутин читают общую очередь; у каждой собственный выход.
	if workers <= 0 {
		return []<-chan StockStatus{}
	}
	outs := make([]<-chan StockStatus, 0, workers)
	for i := 0; i < workers; i++ {
		out := make(chan StockStatus)
		outs = append(outs, out)
		go func() {
			defer close(out)
			for item := range in {
				out <- work(item)
			}
		}()
	}
	return outs
}
func FanIn(streams ...<-chan StockStatus) <-chan StockStatus {
	out := make(chan StockStatus)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan StockStatus) {
			defer wg.Done()
			// TODO: пересылайте сообщения stream в out, пока stream не закроется.
			for status := range stream {
				out <- status
			}
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
func WaitForRestock(ctx context.Context, ready <-chan StockStatus, timeout time.Duration) (StockStatus, error) {
	if timeout <= 0 {
		return StockStatus{}, ErrInvalidTimeout
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return StockStatus{}, ctx.Err()
	case status, ok := <-ready:
		if !ok {
			return StockStatus{}, ErrClosed
		}
		return status, nil
	case <-timer.C:
		return StockStatus{}, ErrTimeout
	}
}
