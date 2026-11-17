package domain

import (
	"errors"
	"time"
)

type Status string

const (
	StatusCreated      Status = "created"
	StatusConfirmed    Status = "confirmed"
	StatusPendingStock Status = "pending_stock"
	StatusCancelled    Status = "cancelled"
)

var (
	ErrItemsMissing    = errors.New("order must have at least one item")
	ErrInvalidQuantity = errors.New("item quantity must be positive")
	ErrInvalidPrice    = errors.New("item price must be positive")
	ErrOrderLimit      = errors.New("order total exceeds limit for this customer")
	ErrPaymentFraud    = errors.New("payment looks fraudulent")
	ErrNotFound        = errors.New("order not found")
	ErrOrderCancelled  = errors.New("cannot update cancelled order")
	ErrInvalidStatus   = errors.New("invalid order status transition")
)

type Item struct {
	SKU      string
	Quantity int
	Price    float64
}

type Order struct {
	ID           int64
	CustomerID   int64
	Items        []Item
	Total        float64
	ConfirmedQty int
	Status       Status
	CreatedAt    time.Time
}

func ValidateItemsPresent(items []Item) error {
	if len(items) == 0 {
		return ErrItemsMissing
	}
	return nil
}

func ValidateItems(items []Item) error {
	if err := ValidateItemsPresent(items); err != nil {
		return err
	}
	for _, item := range items {
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.Price <= 0 {
			return ErrInvalidPrice
		}
	}
	return nil
}

func NewOrder(customer Customer, items []Item) (*Order, error) {
	if err := ValidateItems(items); err != nil {
		return nil, err
	}
	total := itemsTotal(items)
	if total > customer.OrderLimit() {
		return nil, ErrOrderLimit
	}
	return &Order{
		CustomerID: customer.ID,
		Items:      append([]Item(nil), items...),
		Total:      total,
		Status:     StatusCreated,
		CreatedAt:  time.Now(),
	}, nil
}

func (o *Order) ConfirmStock() error {
	if o.Status != StatusCreated {
		return ErrInvalidStatus
	}
	o.Status = StatusConfirmed
	o.ConfirmedQty = itemsQuantity(o.Items)
	return nil
}

func (o *Order) MarkPendingStock() error {
	if o.Status != StatusCreated && o.Status != StatusConfirmed {
		return ErrInvalidStatus
	}
	o.Status = StatusPendingStock
	return nil
}

func (o *Order) ChangeItems(items []Item) error {
	if err := ValidateItemsPresent(items); err != nil {
		return err
	}
	if o.Status == StatusCancelled {
		return ErrOrderCancelled
	}
	if o.Status == StatusConfirmed && itemsQuantity(items) > o.ConfirmedQty {
		if err := o.MarkPendingStock(); err != nil {
			return err
		}
	}
	o.Items = append([]Item(nil), items...)
	o.Total = itemsTotal(items)
	return nil
}

func itemsTotal(items []Item) float64 {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}

func itemsQuantity(items []Item) int {
	var quantity int
	for _, item := range items {
		quantity += item.Quantity
	}
	return quantity
}
