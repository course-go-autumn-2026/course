//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import (
	"errors"
	"fmt"
	"maps"
)

type Store struct {
	balances map[string]int // Условные целые единицы, без дробных сумм.
}

// Unit of Work хранит рабочую копию изменений и применяет их вместе.
// Это модель в памяти для одной горутины, без долговечности и изоляции БД.
// В реальном приложении эту границу обычно обеспечивает транзакция БД.
type UnitOfWork struct {
	store    *Store
	balances map[string]int
	closed   bool
}

func (s *Store) Begin() *UnitOfWork {
	return &UnitOfWork{store: s, balances: maps.Clone(s.balances)}
}

func (u *UnitOfWork) ChangeBalance(account string, delta int) error {
	if u.closed {
		return errors.New("единица работы уже завершена")
	}
	balance, exists := u.balances[account]
	if !exists {
		return fmt.Errorf("счёт %q не найден", account)
	}
	if balance+delta < 0 {
		return errors.New("недостаточно средств")
	}
	u.balances[account] = balance + delta
	return nil
}

func (u *UnitOfWork) Commit() error {
	if u.closed {
		return errors.New("единица работы уже завершена")
	}
	u.store.balances = u.balances
	u.closed = true
	return nil
}

func (u *UnitOfWork) Rollback() {
	u.closed = true // До Commit исходное хранилище не менялось.
}

func Transfer(store *Store, from, to string, amount int) error {
	if amount <= 0 {
		return errors.New("сумма должна быть положительной")
	}
	u := store.Begin()
	defer u.Rollback() // После Commit ничего не меняет.
	if err := u.ChangeBalance(from, -amount); err != nil {
		return err
	}
	if err := u.ChangeBalance(to, amount); err != nil {
		return err // Списание из рабочей копии тоже будет отменено.
	}
	return u.Commit()
}

func main() {
	store := &Store{balances: map[string]int{"Анна": 100, "Борис": 20}}
	err := Transfer(store, "Анна", "Неизвестный", 30)
	fmt.Println("Неудачный перевод:", err)
	fmt.Printf("После отката: Анна=%d, Борис=%d\n", store.balances["Анна"], store.balances["Борис"])

	if err := Transfer(store, "Анна", "Борис", 30); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Printf("После commit: Анна=%d, Борис=%d\n", store.balances["Анна"], store.balances["Борис"])
}
