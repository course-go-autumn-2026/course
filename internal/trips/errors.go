package trips

import "errors"

var (
	ErrNotFound            = errors.New("trip not found")
	ErrCompleted           = errors.New("trip completed")
	ErrDriverBusy          = errors.New("driver busy")
	ErrIdempotencyConflict = errors.New("idempotency key was used with a different request")
)
