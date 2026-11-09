package kitchen

import "time"

// Последовательный код до рефакторинга. Новые функции его не вызывают.
func legacyBakeAll(items []Order, work Baker) []BakedPizza {
	out := make([]BakedPizza, 0, len(items))
	for _, item := range items {
		out = append(out, work(item))
	}
	return out
}
func legacyFanPipeline(in <-chan Order, work Baker) []BakedPizza {
	out := []BakedPizza{}
	for item := range in {
		out = append(out, work(item))
	}
	return out
}
func legacyHeartbeat(repeats int, interval time.Duration, report func()) {
	for i := 0; i < repeats; i++ {
		time.Sleep(interval)
		report()
	}
}
