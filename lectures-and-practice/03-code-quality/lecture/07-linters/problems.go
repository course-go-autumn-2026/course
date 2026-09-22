//go:build lintdemo

package linterdemo

import (
	"fmt"
	"os"
)

// Этот файл намеренно содержит проблемы. Запуск:
// golangci-lint run --build-tags=lintdemo ./07-linters
func readConfig(filename string) string {
	data, _ := os.ReadFile(filename)
	return fmt.Sprintf("%s", data)
}

func unusedHelper() int {
	value := 1
	value = 2
	return value
}
