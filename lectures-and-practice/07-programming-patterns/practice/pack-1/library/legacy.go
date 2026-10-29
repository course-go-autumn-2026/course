package library

// Рабочий код до рефакторинга. Не изменяйте этот файл.

func legacyLoan(kind string, scanning, protectiveCover bool, tariff string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "standard":
		name, price = "StandardLoan", 450
	case "extended":
		name, price = "ExtendedLoan", 550
	default:
		return "", 0, ErrUnknownLoanPlan
	}
	if scanning {
		name += " + scanning"
		price += 50
	}
	if protectiveCover {
		name += " + protective cover"
		price += 30
	}
	switch tariff {
	case "desk_pickup":
	case "home_delivery":
		price += 100
	case "priority_delivery":
		price += 200
	default:
		return "", 0, ErrUnknownTariff
	}
	return name, price, nil
}
