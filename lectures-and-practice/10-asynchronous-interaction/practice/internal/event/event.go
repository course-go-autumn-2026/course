package event

import "errors"

type OrderCreated struct {
	ID      string `json:"id"`
	OrderID string `json:"order_id"`
	Address string `json:"address"`
	Version int    `json:"version"`
}

func (e OrderCreated) Validate() error {
	if e.ID == "" || e.OrderID == "" || e.Address == "" || e.Version != 1 {
		return errors.New("invalid order.created.v1")
	}
	return nil
}
