package pizzeria

import "errors"

// Рабочий код ДО рефакторинга. Файл не изменяют.
// Порядок проверок зашит в if; получатели уведомлений фиксированы;
// настройки передаются двумя позиционными числами без значений по умолчанию.
func legacyValidate(o Order) error {
	if o.Pizzas <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Pizzas > MaxPizzas {
		return ErrApproval
	}
	return nil
}

func legacyReady(o Order, screen, customer Observer) error {
	event := Event{OrderID: o.ID, Status: "ready"}
	screenErr := screen.Update(event)
	customerErr := customer.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewPizzeria(maxPizzas, deliveryFee int) (*Pizzeria, error) {
	if maxPizzas <= 0 {
		return nil, ErrInvalidMax
	}
	if deliveryFee < 0 {
		return nil, ErrInvalidFee
	}
	return &Pizzeria{config: Config{MaxPizzas: maxPizzas, DeliveryFee: deliveryFee}}, nil
}
