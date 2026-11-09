//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import (
	"errors"
	"fmt"
	"strings"
)

var ErrEmptyName = errors.New("имя не заполнено")

type AgeError struct {
	Age int
}

func (e *AgeError) Error() string {
	return fmt.Sprintf("возраст %d вне диапазона 0..150", e.Age)
}

func ValidateProfile(name string, age int) error {
	var problems []error
	if strings.TrimSpace(name) == "" {
		problems = append(problems, ErrEmptyName)
	}
	if age < 0 || age > 150 {
		problems = append(problems, &AgeError{Age: age})
	}
	// Join сохраняет обе ошибки. Если проблем нет, вернётся nil.
	return errors.Join(problems...)
}

func RegisterProfile(name string, age int) error {
	if err := ValidateProfile(name, age); err != nil {
		return fmt.Errorf("регистрация профиля: %w", err)
	}
	fmt.Println("Профиль создан:", name)
	return nil
}

func main() {
	err := RegisterProfile("", -5)
	if err == nil {
		return
	}
	fmt.Println(err)
	fmt.Println("Распознано пустое имя:", errors.Is(err, ErrEmptyName))
	var ageErr *AgeError
	fmt.Println("Распознана ошибка возраста:", errors.As(err, &ageErr))
}
