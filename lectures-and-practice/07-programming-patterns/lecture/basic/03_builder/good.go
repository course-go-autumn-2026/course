//go:build ignore

// Самостоятельный пример. Запуск: go run good.go
// Файлы bad.go и good.go запускаются по отдельности.
// Решение проблемы из bad.go. Сравните зависимости и места изменения правил.

package main

import (
	"errors"
	"fmt"
	"net/url"
)

// Builder собирает адрес по шагам и проверяет обязательное поле в Build.
type URLBuilder struct {
	host  string
	path  string
	query url.Values
}

func NewURLBuilder() *URLBuilder {
	return &URLBuilder{query: make(url.Values)}
}

func (b *URLBuilder) Host(host string) *URLBuilder {
	b.host = host
	return b
}

func (b *URLBuilder) Path(path string) *URLBuilder {
	b.path = path
	return b
}

func (b *URLBuilder) Param(key, value string) *URLBuilder {
	b.query.Set(key, value)
	return b
}

func (b *URLBuilder) Build() (string, error) {
	if b.host == "" {
		return "", errors.New("не указан хост")
	}
	// net/url кодирует пробелы и специальные символы в параметрах.
	u := url.URL{Scheme: "https", Host: b.host, Path: b.path, RawQuery: b.query.Encode()}
	return u.String(), nil
}

func main() {
	address, err := NewURLBuilder().
		Host("example.com").
		Path("/search").
		Param("q", "go patterns").
		Param("page", "1").
		Build()
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println(address)

	_, err = NewURLBuilder().Path("/search").Build()
	fmt.Println("Сборка без хоста:", err)
}
