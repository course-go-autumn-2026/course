package packing

import (
	"context"
	"sync"
	"time"
)

// Задание 1: общая map и Mutex. Каркас ожидания уже дан.
func PackAll(items []PackRequest, work Packer) []PackedOrder {
	results := make(map[int]PackedOrder)
	var wg sync.WaitGroup
	// TODO: объявите общий мьютекс для results.
	for index, item := range items {
		wg.Add(1)
		go func(index int, item PackRequest) {
			defer wg.Done()
			// TODO: вызовите work вне блокировки и сохраните результат под мьютексом.
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
	// TODO: workers горутин читают общую очередь; у каждой собственный выход.
	return nil
}
func FanIn(streams ...<-chan PackedOrder) <-chan PackedOrder {
	out := make(chan PackedOrder)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan PackedOrder) {
			defer wg.Done()
			// TODO: пересылайте сообщения stream в out, пока stream не закроется.
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
func RunPackingReport(ctx context.Context, interval time.Duration, report func()) error {
	return ErrNotImplemented
}
