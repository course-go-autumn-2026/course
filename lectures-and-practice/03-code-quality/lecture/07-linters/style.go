// Пакет linterdemo содержит примеры для демонстрации линтеров и правил оформления кода.
package linterdemo

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrInvalidPort = errors.New("invalid port")

func ParsePort(raw string) (int, error) {
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse port: %w", err)
	}
	if port < 1 || port > 65535 {
		return 0, ErrInvalidPort
	}
	return port, nil
}

func PositiveValues(values []int) []int {
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value > 0 {
			result = append(result, value)
		}
	}
	return result
}
