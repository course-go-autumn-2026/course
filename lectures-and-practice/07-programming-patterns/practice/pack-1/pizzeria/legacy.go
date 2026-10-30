package pizzeria

// Рабочий код ДО рефакторинга. Этот файл не изменяют.
// 1. Выбор пиццы смешан с расчётом заказа.
// 2. Добавки зашиты в булевы флаги и не могут свободно комбинироваться.
// 3. Расчёт заказа сам выбирает доставку по строке.

func legacyOrder(kind string, cheese, box bool, delivery string) (string, int, error) {
	var name string
	var price int
	switch kind {
	case "margherita":
		name, price = "Margherita", 450
	case "pepperoni":
		name, price = "Pepperoni", 550
	default:
		return "", 0, ErrUnknownPizza
	}
	if cheese {
		name += " + cheese"
		price += 50
	}
	if box {
		name += " + box"
		price += 30
	}
	switch delivery {
	case "pickup":
	case "courier":
		price += 100
	case "express":
		price += 200
	default:
		return "", 0, ErrUnknownDelivery
	}
	return name, price, nil
}
