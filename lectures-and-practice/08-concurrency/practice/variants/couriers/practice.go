package couriers

import (
	"context"
	"sync"
	"time"
)

// Задание 1: общая map и Mutex. Каркас ожидания уже дан.
func PrepareAllCouriers(items []CourierTask, work CardBuilder) []CourierCard {
	results := make(map[int]CourierCard)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for index, item := range items {
		wg.Add(1)
		go func(index int, item CourierTask) {
			defer wg.Done()
			// TODO: вызовите work вне блокировки и сохраните результат под мьютексом.
			card := work(item)
			mu.Lock()
			results[index] = card
			mu.Unlock()
		}(index, item)
	}
	wg.Wait()
	output := make([]CourierCard, len(items))
	for index := range items {
		output[index] = results[index]
	}
	return output
}

// Задание 2: Fan-out / Fan-in. Каркас объединения потоков уже дан.
func FanOut(in <-chan CourierTask, workers int, work CardBuilder) []<-chan CourierCard {
	if workers <= 0 {
		return nil
	}
	streams := make([]<-chan CourierCard, 0, workers)
	for i := 0; i < workers; i++ {
		out := make(chan CourierCard)
		streams = append(streams, out)
		go func(out chan<- CourierCard) {
			defer close(out)
			for item := range in {
				out <- work(item)
			}
		}(out)
	}
	return streams
}
func FanIn(streams ...<-chan CourierCard) <-chan CourierCard {
	out := make(chan CourierCard)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan CourierCard) {
			defer wg.Done()
			for card := range stream {
				out <- card
			}
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
func WaitForAcceptance(ctx context.Context, ready <-chan CourierCard, timeout time.Duration) (CourierCard, error) {
	if timeout <= 0 {
		return CourierCard{}, ErrInvalidTimeout
	}
	if err := ctx.Err(); err != nil {
		return CourierCard{}, err
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case card, ok := <-ready:
		if !ok {
			return CourierCard{}, ErrClosed
		}
		return card, nil
	case <-timer.C:
		return CourierCard{}, ErrTimeout
	case <-ctx.Done():
		return CourierCard{}, ctx.Err()
	}
}
