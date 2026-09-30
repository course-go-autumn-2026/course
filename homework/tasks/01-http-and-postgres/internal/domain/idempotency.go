package domain

import "github.com/google/uuid"

type IdempotencyRecord struct {
	RequestHash string
	TripID      uuid.UUID
}
