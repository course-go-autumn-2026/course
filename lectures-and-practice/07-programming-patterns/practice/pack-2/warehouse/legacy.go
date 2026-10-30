package warehouse

// Рабочий код до рефакторинга. Не изменяйте этот файл.
import "errors"

func legacyValidate(o Transfer) error {
	if o.Places <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Places > MaxPlaces {
		return ErrApproval
	}
	return nil
}

func legacyDispatched(o Transfer, operator, client Observer) error {
	event := Event{TransferID: o.ID, Status: "dispatched"}
	screenErr := operator.Update(event)
	customerErr := client.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewWarehouse(limit, fee int) (*Warehouse, error) {
	if limit <= 0 {
		return nil, ErrInvalidMax
	}
	if fee < 0 {
		return nil, ErrInvalidFee
	}
	return &Warehouse{config: Config{MaxPlaces: limit, HandlingFee: fee}}, nil
}
