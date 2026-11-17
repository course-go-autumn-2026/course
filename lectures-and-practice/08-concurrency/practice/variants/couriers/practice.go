package couriers

import (
	"context"
	"sync"
	"time"
)

func PrepareAllCouriers(items []CourierTask, work CardBuilder) []CourierCard {
	results := make(map[int]CourierCard)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for index, item := range items {
		wg.Add(1)
		go func(index int, item CourierTask) {
			defer wg.Done()
			result := work(item)
			mu.Lock()
			results[index] = result
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

func FanOut(in <-chan CourierTask, workers int, work CardBuilder) []<-chan CourierCard {
	if workers <= 0 {
		return nil
	}
	streams := make([]<-chan CourierCard, workers)
	for index := range streams {
		out := make(chan CourierCard)
		streams[index] = out
		go func() {
			defer close(out)
			for item := range in {
				out <- work(item)
			}
		}()
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
			for result := range stream {
				out <- result
			}
		}(stream)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

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
	case result, ok := <-ready:
		if !ok {
			return CourierCard{}, ErrClosed
		}
		return result, nil
	case <-timer.C:
		return CourierCard{}, ErrTimeout
	case <-ctx.Done():
		return CourierCard{}, ctx.Err()
	}
}
