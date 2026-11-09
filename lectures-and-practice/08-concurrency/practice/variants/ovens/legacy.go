package ovens

// Последовательный код до рефакторинга. Новые функции его не вызывают.
func legacyBatch(items []OvenTask, work OvenChecker) []OvenStatus {
	out := make([]OvenStatus, 0, len(items))
	for _, item := range items {
		out = append(out, work(item))
	}
	return out
}
func legacyFanPipeline(in <-chan OvenTask, work OvenChecker) []OvenStatus {
	out := []OvenStatus{}
	for item := range in {
		out = append(out, work(item))
	}
	return out
}
func legacyClock(ready <-chan OvenStatus) OvenStatus { return <-ready }
