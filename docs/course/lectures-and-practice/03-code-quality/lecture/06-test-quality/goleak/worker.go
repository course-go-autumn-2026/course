// Пакет goleakdemo демонстрирует утечку горутины и исправленный вариант.
package goleakdemo

import "context"

func StartLeakyWorker() {
	go func() {
		select {}
	}()
}

func StartWorker(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
	}()
	return done
}
