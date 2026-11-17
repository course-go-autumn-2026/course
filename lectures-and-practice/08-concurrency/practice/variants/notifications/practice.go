package notifications

import (
	"context"
	"sync"
	"time"
)

// Задание 1
func SendAll(items []Notification, work Sender) []DeliveryResult {
	results := make(map[int]DeliveryResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for index, item := range items {
		wg.Add(1)

		go func(index int, item Notification) {
			defer wg.Done()

			result := work(item)

			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(index, item)
	}

	wg.Wait()

	output := make([]DeliveryResult, len(items))
	for index := range items {
		output[index] = results[index]
	}

	return output
}

// Задание 2
func FanOut(in <-chan Notification, workers int, work Sender) []<-chan DeliveryResult {
	if workers <= 0 {
		return nil
	}

	streams := make([]<-chan DeliveryResult, 0, workers)

	for i := 0; i < workers; i++ {
		out := make(chan DeliveryResult)
		streams = append(streams, out)

		go func() {
			defer close(out)

			for item := range in {
				result := work(item)
				out <- result
			}
		}()
	}

	return streams
}

func FanIn(streams ...<-chan DeliveryResult) <-chan DeliveryResult {
	out := make(chan DeliveryResult)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan DeliveryResult) {
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

// Задание 3
func RunNotificationReport(ctx context.Context, interval time.Duration, report func()) error {
	if interval <= 0 {
		return ErrInvalidInterval
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
