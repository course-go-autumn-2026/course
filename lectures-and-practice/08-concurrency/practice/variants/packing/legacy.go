package packing

import "time"

// Последовательный код до рефакторинга. Новые функции его не вызывают.
func legacyBatch(items []PackRequest, work Packer) []PackedOrder {
	out := make([]PackedOrder, 0, len(items))
	for _, item := range items {
		out = append(out, work(item))
	}
	return out
}
func legacyFanPipeline(in <-chan PackRequest, work Packer) []PackedOrder {
	out := []PackedOrder{}
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
