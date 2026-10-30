package fitness

// Рабочий код до рефакторинга. Не изменяйте этот файл.
import "errors"

func legacyValidate(o Visit) error {
	if o.Attendees <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Attendees > MaxAttendees {
		return ErrApproval
	}
	return nil
}

func legacyRegistered(o Visit, operator, client Observer) error {
	event := Event{VisitID: o.ID, Status: "registered"}
	screenErr := operator.Update(event)
	customerErr := client.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewClub(limit, fee int) (*Club, error) {
	if limit <= 0 {
		return nil, ErrInvalidMax
	}
	if fee < 0 {
		return nil, ErrInvalidFee
	}
	return &Club{config: Config{MaxAttendees: limit, RegistrationFee: fee}}, nil
}
