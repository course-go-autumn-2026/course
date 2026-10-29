package hotel

// Рабочий код до рефакторинга. Не изменяйте этот файл.

func legacyBooking(kind string, breakfast, parking bool, tariff string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "standard":
		name, price = "StandardRoom", 450
	case "suite":
		name, price = "SuiteRoom", 550
	default:
		return "", 0, ErrUnknownRoom
	}
	if breakfast {
		name += " + breakfast"
		price += 50
	}
	if parking {
		name += " + parking"
		price += 30
	}
	switch tariff {
	case "direct":
	case "agent":
		price += 100
	case "priority":
		price += 200
	default:
		return "", 0, ErrUnknownTariff
	}
	return name, price, nil
}
