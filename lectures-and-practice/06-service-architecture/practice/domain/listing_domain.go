package listing

import (
	"errors"
	"strings"
	"time"
)

type Status string

type SellerTier string

const(
	TierRegular SellerTier = "regular"
	TierTrusted SellerTier = "trusted"
)

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)


type Seller struct{
	ID int64
	Tier SellerTier
}

type Listing struct {
	ID        int64
	Title     string
	Price     float64
	PhotoURL  string
	SellerID  int64
	Status    Status
	CreatedAt time.Time
}

func NewListing(title string, price float64, photoURL string, sellerID int64) (*Listing, error) {
	if price <= 0 {
		return nil, ErrInvalidPrice
	}
	if strings.TrimSpace(photoURL) == "" {
		return nil, ErrPhotoMissing
	}
	return &Listing{
		Title:     title,
		Price:     price,
		PhotoURL:  photoURL,
		SellerID:  sellerID,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}, nil
}

func (l *Listing) Approve() error {
	if l.Status != StatusPending {
		return errors.New("only pending listing can be approved")
	}
	l.Status = StatusApproved
	return nil
}

func (l *Listing) Reject() error {
	if l.Status != StatusPending {
		return errors.New("only pending listing can be rejected")
	}
	l.Status = StatusRejected
	return nil
}

func ListingValidation(){

}