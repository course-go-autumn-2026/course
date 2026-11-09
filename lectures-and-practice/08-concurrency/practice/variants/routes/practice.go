package routes

import (
	"context"
	"sync"
	"time"
)

// Задание 1: общая map и Mutex. Каркас ожидания уже дан.
func EstimateAll(items []Delivery, work RouteEstimator) []Route {
	results := make(map[int]Route)
	var wg sync.WaitGroup
	// TODO: объявите общий мьютекс для results.
	for index, item := range items {
		wg.Add(1)
		go func(index int, item Delivery) {
			defer wg.Done()
			// TODO: вызовите work вне блокировки и сохраните результат под мьютексом.
		}(index, item)
	}
	wg.Wait()
	output := make([]Route, len(items))
	for index := range items {
		output[index] = results[index]
	}
	return output
}

// Задание 2: Fan-out / Fan-in. Каркас объединения потоков уже дан.
func FanOut(in <-chan Delivery, workers int, work RouteEstimator) []<-chan Route {
	// TODO: workers горутин читают общую очередь; у каждой собственный выход.
	return nil
}
func FanIn(streams ...<-chan Route) <-chan Route {
	out := make(chan Route)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan Route) {
			defer wg.Done()
			// TODO: пересылайте сообщения stream в out, пока stream не закроется.
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
func WaitForCourier(ctx context.Context, ready <-chan Route, timeout time.Duration) (Route, error) {
	return Route{}, ErrNotImplemented
}
