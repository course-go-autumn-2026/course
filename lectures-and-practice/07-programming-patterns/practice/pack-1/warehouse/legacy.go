package warehouse

// Рабочий код до рефакторинга. Не изменяйте этот файл.

func legacyStorage(kind string, climateControl, insurance bool, tariff string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "small":
		name, price = "SmallCell", 450
	case "large":
		name, price = "LargeCell", 550
	default:
		return "", 0, ErrUnknownStorageCell
	}
	if climateControl {
		name += " + climate control"
		price += 50
	}
	if insurance {
		name += " + insurance"
		price += 30
	}
	switch tariff {
	case "self":
	case "standard":
		price += 100
	case "express":
		price += 200
	default:
		return "", 0, ErrUnknownTariff
	}
	return name, price, nil
}
