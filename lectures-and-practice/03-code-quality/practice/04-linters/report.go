package linters

import (
	"encoding/json"
	"fmt"
	"os"
)

type Report struct {
	Title string `json:"title"`
	Total int    `json:"total"`
}

func SaveReport(path string, report Report) {
	data, _ := json.Marshal(report)
	_ = os.WriteFile(path, data, 0o644)
}

func DisplayName(name string) string {
	return fmt.Sprintf("%s", name)
}

func IsReady(ready bool) bool {
	if ready == true {
		return true
	}
	return false
}

func Total(values []int) int {
	total := 100
	total = 0
	for _, value := range values {
		total += value
	}
	return total
}
