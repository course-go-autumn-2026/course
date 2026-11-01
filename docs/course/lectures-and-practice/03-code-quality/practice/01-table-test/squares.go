package tabletest

import (
	"errors"
	"fmt"
	"sync"
)

var ErrNegativeNumber = errors.New("negative number")

// SumSquares параллельно вычисляет квадраты чисел и возвращает их сумму.
func SumSquares(numbers []int) (int, error) {
	results := make(chan int)
	errorsCh := make(chan error, 1)

	var workers sync.WaitGroup
	for _, number := range numbers {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if number < 0 {
				select {
				case errorsCh <- fmt.Errorf("%w: %d", ErrNegativeNumber, number):
				default:
				}
				return
			}
			results <- number * number
		}()
	}

	go func() {
		workers.Wait()
		close(results)
	}()

	total := 0
	for result := range results {
		total += result
	}

	select {
	case err := <-errorsCh:
		return 0, err
	default:
		return total, nil
	}
}
