// Package domain - модели предметной области. Ничего не знает
// про HTTP, SQL или конкретных поставщиков модерации.
package domain

import (
	"errors"
	"strings"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

var (
	ErrInvalidPrice = errors.New("price must be positive")
	ErrPhotoMissing = errors.New("photo is required")
	ErrNotFound     = errors.New("listing not found")
)

// Listing - богатая модель: умеет проверять собственные инварианты
// и менять своё состояние. Сравните с before/handler.go - там те же
// два правила лежали в хендлере.
type Listing struct {
	ID        int64
	Title     string
	Price     float64
	PhotoURL  string
	SellerID  int64
	Status    Status
	CreatedAt time.Time
}

// NewListing - единственный способ создать объявление. Невозможно
// получить Listing с нарушенными инвариантами: конструктор их
// проверяет сразу.
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

// Approve - поведение, а не просто смена поля. Модель сама решает,
// из каких состояний можно перейти в approved.
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
