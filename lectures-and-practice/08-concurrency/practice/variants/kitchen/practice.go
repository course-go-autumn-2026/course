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
	// TODO: объявите общий мьютекс для results.
	for index, item := range items {
		wg.Add(1)
		go func(index int, item Order) {
			defer wg.Done()
			// TODO: вызовите work вне блокировки и сохраните результат под мьютексом.
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
	return nil
}
func FanIn(streams ...<-chan BakedPizza) <-chan BakedPizza {
	out := make(chan BakedPizza)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan BakedPizza) {
			defer wg.Done()
			// TODO: пересылайте сообщения stream в out, пока stream не закроется.
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
func RunHeartbeat(ctx context.Context, interval time.Duration, report func()) error {
	return ErrNotImplemented
}
