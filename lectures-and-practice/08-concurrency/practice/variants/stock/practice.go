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
	// TODO: объявите общий мьютекс для results.
	for index, item := range items {
		wg.Add(1)
		go func(index int, item Ingredient) {
			defer wg.Done()
			// TODO: вызовите work вне блокировки и сохраните результат под мьютексом.
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
	return nil
}
func FanIn(streams ...<-chan StockStatus) <-chan StockStatus {
	out := make(chan StockStatus)
	var wg sync.WaitGroup
	for _, stream := range streams {
		wg.Add(1)
		go func(stream <-chan StockStatus) {
			defer wg.Done()
			// TODO: пересылайте сообщения stream в out, пока stream не закроется.
		}(stream)
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// Задание 3: таймер или тикер и контекст.
// TODO: выполните контракт из README.
func WaitForRestock(ctx context.Context, ready <-chan StockStatus, timeout time.Duration) (StockStatus, error) {
	return StockStatus{}, ErrNotImplemented
}
