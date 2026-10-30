package courses

// Рабочий код до рефакторинга. Не изменяйте этот файл.
import "errors"

func legacyValidate(o Enrollment) error {
	if o.Students <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Students > MaxStudents {
		return ErrApproval
	}
	return nil
}

func legacyActivated(o Enrollment, operator, client Observer) error {
	event := Event{EnrollmentID: o.ID, Status: "activated"}
	screenErr := operator.Update(event)
	customerErr := client.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewSchool(limit, fee int) (*School, error) {
	if limit <= 0 {
		return nil, ErrInvalidMax
	}
	if fee < 0 {
		return nil, ErrInvalidFee
	}
	return &School{config: Config{MaxStudents: limit, EnrollmentFee: fee}}, nil
}
