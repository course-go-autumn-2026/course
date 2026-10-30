package domain

import "errors"

var (
	ErrInvalidPrice         = errors.New("price must be positive")
	ErrPhotoMissing         = errors.New("photo is required")
	ErrPriceSuspicious      = errors.New("price looks suspicious")
	ErrNotFound             = errors.New("listing not found")
	ErrCannotUpdateRejected = errors.New("cannot update rejected listing")
)
