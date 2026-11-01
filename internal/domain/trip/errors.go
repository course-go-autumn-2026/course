package trip

import "errors"

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Reason
}

var (
	ErrNotFound            = errors.New("trip not found")
	ErrCompleted           = errors.New("trip completed")
	ErrDriverBusy          = errors.New("driver busy")
	ErrIdempotencyConflict = errors.New("idempotency key was used with a different request")
)
