package courses

// Рабочий код до рефакторинга. Не изменяйте этот файл.

func legacyEnrollment(kind string, mentoring, certificate bool, tariff string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "recorded":
		name, price = "RecordedCourse", 450
	case "live":
		name, price = "LiveCourse", 550
	default:
		return "", 0, ErrUnknownCourse
	}
	if mentoring {
		name += " + mentoring"
		price += 50
	}
	if certificate {
		name += " + certificate"
		price += 30
	}
	switch tariff {
	case "self":
	case "assisted":
		price += 100
	case "priority":
		price += 200
	default:
		return "", 0, ErrUnknownTariff
	}
	return name, price, nil
}
