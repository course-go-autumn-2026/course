package autoservice

// Рабочий код до рефакторинга. Не изменяйте этот файл.
import "errors"

func legacyValidate(o WorkOrder) error {
	if o.Jobs <= 0 {
		return ErrEmpty
	}
	if o.Total <= 0 {
		return ErrTotal
	}
	if o.Jobs > MaxJobs {
		return ErrApproval
	}
	return nil
}

func legacyCompleted(o WorkOrder, operator, client Observer) error {
	event := Event{WorkOrderID: o.ID, Status: "completed"}
	screenErr := operator.Update(event)
	customerErr := client.Update(event)
	return errors.Join(screenErr, customerErr)
}

func legacyNewWorkshop(limit, fee int) (*Workshop, error) {
	if limit <= 0 {
		return nil, ErrInvalidMax
	}
	if fee < 0 {
		return nil, ErrInvalidFee
	}
	return &Workshop{config: Config{MaxJobs: limit, CalloutFee: fee}}, nil
}
