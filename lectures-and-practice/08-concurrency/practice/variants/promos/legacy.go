package promos

import "time"

// Последовательный код до рефакторинга. Новые функции его не вызывают.
func legacyBatch(items []PromoRequest, work QuoteBuilder) []Quote {
	out := make([]Quote, 0, len(items))
	for _, item := range items {
		out = append(out, work(item))
	}
	return out
}
func legacyFanPipeline(in <-chan PromoRequest, work QuoteBuilder) []Quote {
	out := []Quote{}
	for item := range in {
		out = append(out, work(item))
	}
	return out
}
func legacyClock(repeats int, interval time.Duration, report func()) {
	for i := 0; i < repeats; i++ {
		time.Sleep(interval)
		report()
	}
}
