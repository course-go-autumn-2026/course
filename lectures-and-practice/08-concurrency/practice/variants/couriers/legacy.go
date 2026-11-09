package couriers

// Последовательный код до рефакторинга. Новые функции его не вызывают.
func legacyBatch(items []CourierTask, work CardBuilder) []CourierCard {
	out := make([]CourierCard, 0, len(items))
	for _, item := range items {
		out = append(out, work(item))
	}
	return out
}
func legacyFanPipeline(in <-chan CourierTask, work CardBuilder) []CourierCard {
	out := []CourierCard{}
	for item := range in {
		out = append(out, work(item))
	}
	return out
}
func legacyClock(ready <-chan CourierCard) CourierCard { return <-ready }
