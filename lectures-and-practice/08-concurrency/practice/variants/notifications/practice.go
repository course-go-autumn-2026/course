package notifications

import (
	"context"
	"sync"
	"time"
)

// Задание 1: общая map и Mutex. Каркас ожидания уже дан.
func SendAll(items []Notification, work Sender) []DeliveryResult {
	results := make(map[int]DeliveryResult)
	var wg sync.WaitGroup
	// TODO: объявите общий мьютекс для results.
	var mu sync.Mutex
	for index, item := range items {
		wg.Add(1)
		go func(index int, item Notification) {
			defer wg.Done()
			// TODO: вызовите work вне блокировки и сохраните результат под мьютексом.
			out := work(item)
			mu.Lock()
			results[index] = out
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

// Задание 2: Fan-out / Fan-in. Каркас объединения потоков уже дан.
func FanOut(in <-chan Notification, workers int, work Sender) []<-chan DeliveryResult {
	// TODO: workers горутин читают общую очередь; у каждой собственный выход.
	if workers <= 0 {
		return nil
	}

	streams := make([]<-chan DeliveryResult, workers)
	for index := range streams {
		out := make(chan DeliveryResult)
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
func FanIn(streams ...<-chan DeliveryResult) <-chan DeliveryResult {
	out := make(chan DeliveryResult)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan DeliveryResult) {
			defer wg.Done()
			// TODO: пересылайте сообщения stream в out, пока stream не закроется.
			for result := range stream {
				out <- result
			}
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
// Реализация: тикер с interval вызывает report до отмены контекста.
func RunNotificationReport(ctx context.Context, interval time.Duration, report func()) error {
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
