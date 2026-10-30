package main

import (
	"errors"
	"strings"
)

type Order struct {
	ItemsCount int
	Address    string
	Paid       bool
}

// Текущие требования: отправить можно только оплаченный заказ
// с хотя бы одной позицией и непустым адресом.
// Если нарушено несколько условий, возвращаем первую ошибку в этом порядке.
func BeforeValidateDispatch(order Order) error {
	if order.ItemsCount > 0 {
		if strings.TrimSpace(order.Address) != "" {
			if order.Paid {
				return nil
			} else {
				return errors.New("заказ не оплачен")
			}
		} else {
			return errors.New("не указан адрес")
		}
	} else {
		return errors.New("заказ пуст")
	}
}
