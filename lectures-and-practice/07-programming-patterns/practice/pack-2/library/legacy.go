package library

// Рабочий код до рефакторинга. Не изменяйте этот файл.
import "errors"

func legacyValidate(o LoanRequest) error {
	if o.Copies <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Copies > MaxCopies {
		return ErrApproval
	}
	return nil
}

func legacyIssued(o LoanRequest, operator, client Observer) error {
	event := Event{LoanRequestID: o.ID, Status: "issued"}
	screenErr := operator.Update(event)
	customerErr := client.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewLibrary(limit, fee int) (*Library, error) {
	if limit <= 0 {
		return nil, ErrInvalidMax
	}
	if fee < 0 {
		return nil, ErrInvalidFee
	}
	return &Library{config: Config{MaxCopies: limit, DeliveryFee: fee}}, nil
}
