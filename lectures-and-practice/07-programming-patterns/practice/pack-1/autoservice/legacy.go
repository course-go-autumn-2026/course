package autoservice

// Рабочий код до рефакторинга. Не изменяйте этот файл.

func legacyService(kind string, carWash, prioritySlot bool, tariff string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "inspection":
		name, price = "Inspection", 450
	case "alignment":
		name, price = "WheelAlignment", 550
	default:
		return "", 0, ErrUnknownService
	}
	if carWash {
		name += " + car wash"
		price += 50
	}
	if prioritySlot {
		name += " + priority slot"
		price += 30
	}
	switch tariff {
	case "scheduled":
	case "same_day":
		price += 100
	case "night_shift":
		price += 200
	default:
		return "", 0, ErrUnknownTariff
	}
	return name, price, nil
}
