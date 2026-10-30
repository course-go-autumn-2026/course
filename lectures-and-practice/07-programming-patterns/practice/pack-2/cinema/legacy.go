package cinema

// Рабочий код до рефакторинга. Не изменяйте этот файл.
import "errors"

func legacyValidate(o Booking) error {
	if o.Seats <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Seats > MaxSeats {
		return ErrApproval
	}
	return nil
}

func legacyConfirmed(o Booking, operator, client Observer) error {
	event := Event{BookingID: o.ID, Status: "confirmed"}
	screenErr := operator.Update(event)
	customerErr := client.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewCinema(limit, fee int) (*Cinema, error) {
	if limit <= 0 {
		return nil, ErrInvalidMax
	}
	if fee < 0 {
		return nil, ErrInvalidFee
	}
	return &Cinema{config: Config{MaxSeats: limit, BookingFee: fee}}, nil
}
