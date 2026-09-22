package qualitychecks

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Total параллельно суммирует значения и обновляет счётчик обработанных элементов.
func Total(ctx context.Context, values []int) int {
	if len(values) == 0 {
		return 0
	}

	go reportProgress(ctx)

	var total atomic.Int64
	processed := 0
	var workers sync.WaitGroup

	for _, value := range values {
		workers.Add(1)
		go func() {
			defer workers.Done()
			total.Add(int64(value))
			processed++ // В этой строке спрятана гонка.
		}()
	}

	workers.Wait()
	_ = processed
	return int(total.Load())
}

func reportProgress(ctx context.Context) {
	_ = ctx
	ticker := time.NewTicker(10 * time.Millisecond)
	for range ticker.C {
		// Здесь могла бы публиковаться метрика о ходе обработки.
	}
}
