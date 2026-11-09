//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: два изменения выполняются независимо; при ошибке деньги пропадают.
// Unit of Work применит их вместе либо отбросит всю рабочую копию.
// Оба примера работают только в памяти, без настоящих платежей.

package main

import (
	"errors"
	"fmt"
)

type Store struct{ balances map[string]int }

func (s *Store) ChangeBalance(account string, delta int) error {
	balance, exists := s.balances[account]
	if !exists {
		return fmt.Errorf("счёт %q не найден", account)
	}
	if balance+delta < 0 {
		return errors.New("недостаточно средств")
	}
	s.balances[account] = balance + delta
	return nil
}

func Transfer(store *Store, from, to string, amount int) error {
	if amount <= 0 {
		return errors.New("сумма должна быть положительной")
	}
	if err := store.ChangeBalance(from, -amount); err != nil {
		return err
	}
	// Если зачисление не удалось, списание уже осталось в хранилище.
	return store.ChangeBalance(to, amount)
}

func main() {
	store := &Store{balances: map[string]int{"Анна": 100, "Борис": 20}}
	err := Transfer(store, "Анна", "Неизвестный", 30)
	fmt.Println("Неудачный перевод:", err)
	fmt.Printf("После ошибки: Анна=%d, Борис=%d\n", store.balances["Анна"], store.balances["Борис"])

	if err := Transfer(store, "Анна", "Борис", 30); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Printf("После перевода: Анна=%d, Борис=%d\n", store.balances["Анна"], store.balances["Борис"])
}
