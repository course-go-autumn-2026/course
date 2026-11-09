package receipts

import "time"

// Последовательный код до рефакторинга. Новые функции его не вызывают.
func legacyBatch(items []ReceiptRequest, work ReceiptPrinter) []Receipt {
	out := make([]Receipt, 0, len(items))
	for _, item := range items {
		out = append(out, work(item))
	}
	return out
}
func legacyFanPipeline(in <-chan ReceiptRequest, work ReceiptPrinter) []Receipt {
	out := []Receipt{}
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
