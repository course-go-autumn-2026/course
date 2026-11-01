// Пакет coveragedemo содержит небольшой пример для демонстрации покрытия кода.
package coveragedemo

func ShippingCost(weightKG int, express bool) int {
	if weightKG <= 0 {
		return 0
	}

	cost := 300
	if weightKG > 10 {
		cost += (weightKG - 10) * 25
	}
	if express {
		cost *= 2
	}

	return cost
}
