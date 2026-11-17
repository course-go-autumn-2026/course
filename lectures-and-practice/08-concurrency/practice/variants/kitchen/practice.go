package kitchen

import (
	"context"
	"sync"
	"time"
)

// Задание 1: общая map и Mutex. Каркас ожидания уже дан.
func BakeAll(items []Order, work Baker) []BakedPizza {
	results := make(map[int]BakedPizza)
	var wg sync.WaitGroup
	var mtx sync.Mutex

	for index, item := range items {
		wg.Add(1)
		go func(index int, item Order) {
			defer wg.Done()
			workResponse := work(item)

			mtx.Lock()
			results[index] = workResponse
			mtx.Unlock()
		}(index, item)
	}

	wg.Wait()
	output := make([]BakedPizza, len(items))
	for index := range items {
		output[index] = results[index]
	}
	return output
}

// Задание 2: Fan-out / Fan-in. Каркас объединения потоков уже дан.
func FanOut(in <-chan Order, workers int, work Baker) []<-chan BakedPizza {
	// TODO: workers горутин читают общую очередь; у каждой собственный выход.
	if workers <= 0 {
		return nil
	}
	workersArr := make([]chan BakedPizza, workers)
	for i := range workers {
		workersArr[i] = make(chan BakedPizza)
	}

	for i := range workers {
		go func() {
			defer close(workersArr[i])
			for order := range in {
				workersArr[i] <- work(order)
			}
		}()
	}

	result := make([]<-chan BakedPizza, workers)
	for i := range workersArr {
		result[i] = workersArr[i]
	}

	return result
}
func FanIn(streams ...<-chan BakedPizza) <-chan BakedPizza {
	out := make(chan BakedPizza)
	var wg sync.WaitGroup

	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan BakedPizza) {
			defer wg.Done()
			for pizza := range stream {
				out <- pizza
			}
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
func RunHeartbeat(ctx context.Context, interval time.Duration, report func()) error {
	if interval <= 0 {
		return ErrInvalidInterval
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			report()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
