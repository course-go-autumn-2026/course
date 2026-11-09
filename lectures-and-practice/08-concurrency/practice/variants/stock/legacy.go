package stock

// Последовательный код до рефакторинга. Новые функции его не вызывают.
func legacyBatch(items []Ingredient, work StockChecker) []StockStatus {
	out := make([]StockStatus, 0, len(items))
	for _, item := range items {
		out = append(out, work(item))
	}
	return out
}
func legacyFanPipeline(in <-chan Ingredient, work StockChecker) []StockStatus {
	out := []StockStatus{}
	for item := range in {
		out = append(out, work(item))
	}
	return out
}
func legacyClock(ready <-chan StockStatus) StockStatus { return <-ready }
