// Пакет tabletestdemo демонстрирует табличные тесты и проверки с testify.
package tabletestdemo

import (
	"errors"
	"slices"
	"strings"
)

var ErrEmptyTags = errors.New("tags are empty")

// NormalizeTags обрезает пробелы, приводит теги к нижнему регистру, удаляет дубликаты и сортирует результат.
func NormalizeTags(tags []string) ([]string, error) {
	if len(tags) == 0 {
		return nil, ErrEmptyTags
	}

	unique := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" {
			unique[tag] = struct{}{}
		}
	}

	result := make([]string, 0, len(unique))
	for tag := range unique {
		result = append(result, tag)
	}
	slices.Sort(result)

	return result, nil
}
