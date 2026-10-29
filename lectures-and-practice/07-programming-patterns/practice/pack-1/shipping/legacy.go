package shipping

// Рабочий код до рефакторинга. Не изменяйте этот файл.

func legacyShipment(kind string, insurance, fragileHandling bool, tariff string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "envelope":
		name, price = "Envelope", 450
	case "parcel":
		name, price = "Parcel", 550
	default:
		return "", 0, ErrUnknownShipment
	}
	if insurance {
		name += " + insurance"
		price += 50
	}
	if fragileHandling {
		name += " + fragile handling"
		price += 30
	}
	switch tariff {
	case "drop_off":
	case "courier_pickup":
		price += 100
	case "express_pickup":
		price += 200
	default:
		return "", 0, ErrUnknownTariff
	}
	return name, price, nil
}
