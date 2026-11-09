package notifications

import "time"

// Последовательный код до рефакторинга. Новые функции его не вызывают.
func legacyBatch(items []Notification, work Sender) []DeliveryResult {
	out := make([]DeliveryResult, 0, len(items))
	for _, item := range items {
		out = append(out, work(item))
	}
	return out
}
func legacyFanPipeline(in <-chan Notification, work Sender) []DeliveryResult {
	out := []DeliveryResult{}
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
