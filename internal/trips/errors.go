package trips

import "errors"

var (
	ErrNotFound   = errors.New("trip not found")
	ErrCompleted  = errors.New("trip completed")
	ErrDriverBusy = errors.New("driver busy")
)
