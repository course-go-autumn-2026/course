package main

import (
	"errors"
	"strings"
)

// Те же требования и порядок ошибок, но меньше вложенности.
// Упрощение не убирает проверки и не меняет правила отправки.
func ValidateDispatch(order Order) error {
	if order.ItemsCount <= 0 {
		return errors.New("заказ пуст")
	}

	if strings.TrimSpace(order.Address) == "" {
		return errors.New("не указан адрес")
	}

	if !order.Paid {
		return errors.New("заказ не оплачен")
	}

	return nil
}
