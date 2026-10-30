package hotel

// Рабочий код до рефакторинга. Не изменяйте этот файл.
import "errors"

func legacyValidate(o Reservation) error {
	if o.Nights <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Nights > MaxNights {
		return ErrApproval
	}
	return nil
}

func legacyConfirmed(o Reservation, operator, client Observer) error {
	event := Event{ReservationID: o.ID, Status: "confirmed"}
	screenErr := operator.Update(event)
	customerErr := client.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewHotel(limit, fee int) (*Hotel, error) {
	if limit <= 0 {
		return nil, ErrInvalidMax
	}
	if fee < 0 {
		return nil, ErrInvalidFee
	}
	return &Hotel{config: Config{MaxNights: limit, ServiceFee: fee}}, nil
}
