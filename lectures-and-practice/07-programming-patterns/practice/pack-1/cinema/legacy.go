package cinema

// Рабочий код до рефакторинга. Не изменяйте этот файл.

func legacyTicket(kind string, popcorn, drink bool, tariff string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "standard":
		name, price = "StandardTicket", 450
	case "premium":
		name, price = "PremiumTicket", 550
	default:
		return "", 0, ErrUnknownTicket
	}
	if popcorn {
		name += " + popcorn"
		price += 50
	}
	if drink {
		name += " + drink"
		price += 30
	}
	switch tariff {
	case "box_office":
	case "online":
		price += 100
	case "concierge":
		price += 200
	default:
		return "", 0, ErrUnknownTariff
	}
	return name, price, nil
}
