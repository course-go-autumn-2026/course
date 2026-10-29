package shop

// Рабочий код до рефакторинга. Не изменяйте этот файл.
import "errors"

func legacyValidate(o Purchase) error {
	if o.Items <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Items > MaxItems {
		return ErrApproval
	}
	return nil
}

func legacyShipped(o Purchase, operator, client Observer) error {
	event := Event{PurchaseID: o.ID, Status: "shipped"}
	screenErr := operator.Update(event)
	customerErr := client.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewShop(limit, fee int) (*Shop, error) {
	if limit <= 0 {
		return nil, ErrInvalidMax
	}
	if fee < 0 {
		return nil, ErrInvalidFee
	}
	return &Shop{config: Config{MaxItems: limit, DeliveryFee: fee}}, nil
}
