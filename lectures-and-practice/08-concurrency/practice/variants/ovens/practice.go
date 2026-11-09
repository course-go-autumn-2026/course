package ovens

import (
	"context"
	"sync"
	"time"
)

// Задание 1: общая map и Mutex. Каркас ожидания уже дан.
func CheckAllOvens(items []OvenTask, work OvenChecker) []OvenStatus {
	results := make(map[int]OvenStatus)
	var wg sync.WaitGroup
	// TODO: объявите общий мьютекс для results.
	for index, item := range items {
		wg.Add(1)
		go func(index int, item OvenTask) {
			defer wg.Done()
			// TODO: вызовите work вне блокировки и сохраните результат под мьютексом.
		}(index, item)
	}
	wg.Wait()
	output := make([]OvenStatus, len(items))
	for index := range items {
		output[index] = results[index]
	}
	return output
}

// Задание 2: Fan-out / Fan-in. Каркас объединения потоков уже дан.
func FanOut(in <-chan OvenTask, workers int, work OvenChecker) []<-chan OvenStatus {
	// TODO: workers горутин читают общую очередь; у каждой собственный выход.
	return nil
}
func FanIn(streams ...<-chan OvenStatus) <-chan OvenStatus {
	out := make(chan OvenStatus)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan OvenStatus) {
			defer wg.Done()
			// TODO: пересылайте сообщения stream в out, пока stream не закроется.
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
func WaitForOven(ctx context.Context, ready <-chan OvenStatus, timeout time.Duration) (OvenStatus, error) {
	return OvenStatus{}, ErrNotImplemented
}
