package fitness

// Рабочий код до рефакторинга. Не изменяйте этот файл.

func legacyMembership(kind string, locker, towel bool, tariff string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "daytime":
		name, price = "DaytimePass", 450
	case "full_day":
		name, price = "FullDayPass", 550
	default:
		return "", 0, ErrUnknownMembership
	}
	if locker {
		name += " + locker"
		price += 50
	}
	if towel {
		name += " + towel"
		price += 30
	}
	switch tariff {
	case "next_week":
	case "today":
		price += 100
	case "personal":
		price += 200
	default:
		return "", 0, ErrUnknownTariff
	}
	return name, price, nil
}
