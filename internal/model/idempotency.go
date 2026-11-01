package model

import (
	"time"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
)

type IdempotencyRecord struct {
	RequestHash []byte
	Trip        trip.Trip
	ExpiresAt   time.Time
}
