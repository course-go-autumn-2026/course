package trip

import (
	"errors"
	"fmt"
)

var (
	ErrTripNotFound  = errors.New("trip not found")
	ErrDriverBusy    = errors.New("driver already has an active trip")
	ErrTripCompleted = errors.New("trip is already completed")
)

type ValidationError struct {
	Field   string
	Message string
}

func (errorValue *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", errorValue.Field, errorValue.Message)
}
