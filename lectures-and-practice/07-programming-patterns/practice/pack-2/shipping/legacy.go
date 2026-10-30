package shipping

// Рабочий код до рефакторинга. Не изменяйте этот файл.
import "errors"

func legacyValidate(o Dispatch) error {
	if o.Packages <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Packages > MaxPackages {
		return ErrApproval
	}
	return nil
}

func legacyDispatched(o Dispatch, operator, client Observer) error {
	event := Event{DispatchID: o.ID, Status: "dispatched"}
	screenErr := operator.Update(event)
	customerErr := client.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewDeliveryService(limit, fee int) (*DeliveryService, error) {
	if limit <= 0 {
		return nil, ErrInvalidMax
	}
	if fee < 0 {
		return nil, ErrInvalidFee
	}
	return &DeliveryService{config: Config{MaxPackages: limit, PickupFee: fee}}, nil
}
