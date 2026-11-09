package routes

// Последовательный код до рефакторинга. Новые функции его не вызывают.
func legacyBatch(items []Delivery, work RouteEstimator) []Route {
	out := make([]Route, 0, len(items))
	for _, item := range items {
		out = append(out, work(item))
	}
	return out
}
func legacyFanPipeline(in <-chan Delivery, work RouteEstimator) []Route {
	out := []Route{}
	for item := range in {
		out = append(out, work(item))
	}
	return out
}
func legacyClock(ready <-chan Route) Route { return <-ready }
