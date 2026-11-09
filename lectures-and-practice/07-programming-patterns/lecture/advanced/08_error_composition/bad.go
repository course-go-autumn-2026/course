//go:build ignore

// Самостоятельный пример. Запуск: go run bad.go
// Файлы bad.go и good.go запускаются по отдельности.
// Проблема: текст есть, а распознать причины программно через Is/As уже нельзя.
// errors.Join и обёртка %w сохранят исходные значения ошибок.

package main

import (
	"errors"
	"fmt"
	"strings"
)

var ErrEmptyName = errors.New("имя не заполнено")

type AgeError struct{ Age int }

func (e *AgeError) Error() string {
	return fmt.Sprintf("возраст %d вне диапазона 0..150", e.Age)
}

func ValidateProfile(name string, age int) error {
	var messages []string
	if strings.TrimSpace(name) == "" {
		messages = append(messages, ErrEmptyName.Error())
	}
	if age < 0 || age > 150 {
		messages = append(messages, (&AgeError{Age: age}).Error())
	}
	if len(messages) == 0 {
		return nil
	}
	// Исходные значения ошибок потеряны, остался только текст.
	return errors.New(strings.Join(messages, "\n"))
}

func RegisterProfile(name string, age int) error {
	if err := ValidateProfile(name, age); err != nil {
		return fmt.Errorf("регистрация профиля: %v", err) // Ещё одна потеря цепочки.
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
