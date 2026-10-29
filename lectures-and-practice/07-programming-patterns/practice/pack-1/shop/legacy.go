package shop

// Рабочий код до рефакторинга. Не изменяйте этот файл.

func legacyPurchase(kind string, warranty, giftWrap bool, tariff string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "keyboard":
		name, price = "Keyboard", 450
	case "mouse":
		name, price = "Mouse", 550
	default:
		return "", 0, ErrUnknownProduct
	}
	if warranty {
		name += " + warranty"
		price += 50
	}
	if giftWrap {
		name += " + gift wrap"
		price += 30
	}
	switch tariff {
	case "store_pickup":
	case "courier":
		price += 100
	case "express":
		price += 200
	default:
		return "", 0, ErrUnknownTariff
	}
	return name, price, nil
}
