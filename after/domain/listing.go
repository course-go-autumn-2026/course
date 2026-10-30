package domain

import (
	"strings"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
	StatusReview   Status = "review"
)

// богатая модель.
type Listing struct {
	ID        int64
	Title     string
	Price     float64
	PhotoURL  string
	SellerID  int64
	Status    Status
	CreatedAt time.Time
}

// хардкод порогов из ТЗ
func priceLimit(sellerID int64) float64 {
	if sellerID == 2 {
		return 50_000_000.0
	}
	return 10_000_000.0
}

// конструктор с проверками правил 1-3
func NewListing(title string, price float64, photoURL string, sellerID int64) (*Listing, error) {
	if price <= 0 {
		return nil, ErrInvalidPrice
	}
	if strings.TrimSpace(photoURL) == "" {
		return nil, ErrPhotoMissing
	}
	if price > priceLimit(sellerID) {
		return nil, ErrPriceSuspicious
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

// правила 4 и 5
func (l *Listing) ChangePrice(newPrice float64) error {
	if newPrice <= 0 {
		return ErrInvalidPrice
	}
	if l.Status == StatusRejected {
		return ErrCannotUpdateRejected
	}
	l.Price = newPrice
	if l.Status == StatusApproved && newPrice > priceLimit(l.SellerID) {
		l.Status = StatusPending
	}
	return nil
}

// Переходы состояний
func (l *Listing) Approve()     { l.Status = StatusApproved }
func (l *Listing) Reject()      { l.Status = StatusRejected }
func (l *Listing) MarkReview()  { l.Status = StatusReview }
func (l *Listing) MarkPending() { l.Status = StatusPending }
