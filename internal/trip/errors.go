package trip

import "errors"

var (
	ErrTripNotFound  = errors.New("trip not found")
	ErrDriverBusy    = errors.New("driver already has an active trip")
	ErrTripCompleted = errors.New("trip is already completed")
)
