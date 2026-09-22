// Пакет flakydemo содержит намеренно неоднозначную функцию для демонстрации нестабильного теста.
package flakydemo

func MaxValueKey(values map[string]int) string {
	maxValue := -int(^uint(0)>>1) - 1
	var maxKey string
	for key, value := range values {
		if value > maxValue {
			maxValue = value
			maxKey = key
		}
	}
	return maxKey
}
