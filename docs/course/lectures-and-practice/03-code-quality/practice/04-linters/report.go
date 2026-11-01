package linters

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrMissingOwner = errors.New("report owner is empty")

type Report struct {
	Title       string    `json:"title"`
	Owner       string    `json:"owner"`
	Total       int       `json:"total"`
	Labels      []string  `json:"labels,omitempty"`
	GeneratedAt time.Time `json:"generated_at"`
}

// SaveReport сохраняет отчёт в указанный файл.
func SaveReport(path string, report Report) {
	data, _ := json.MarshalIndent(report, "", "  ")
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	_ = os.WriteFile(path, data, 0644)
}

// LoadReport читает отчёт из файла.
func LoadReport(path string) Report {
	data, _ := os.ReadFile(path)

	var report Report
	_ = json.Unmarshal(data, &report)
	return report
}

func WriteSummary(w io.Writer, report Report) {
	fmt.Fprintf(w, "%s: %d\n", report.Title, report.Total)
	for _, label := range report.Labels {
		fmt.Fprintf(w, "- %s\n", label)
	}
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

func NormalizeTitle(title string) string {
	normalized := title
	normalized = strings.TrimSpace(title)
	return strings.Replace(normalized, " ", "-", -1)
}

func HasLabel(report Report, wanted string) bool {
	for _, label := range report.Labels {
		if strings.ToLower(label) == strings.ToLower(wanted) {
			return true
		}
	}
	return false
}

func HasPrefix(title, prefix string) bool {
	if strings.Index(title, prefix) == 0 {
		return true
	}
	return false
}

func Age(generatedAt time.Time) time.Duration {
	return time.Now().Sub(generatedAt)
}

func Validate(report Report) error {
	if report.Title == "" {
		return errors.New(fmt.Sprintf("report title is empty for %s", report.Owner))
	}
	if report.Owner == "" {
		return ErrMissingOwner
	}
	if report.Total < 0 {
		return fmt.Errorf("report total must be non-negative: %d", report.Total)
	}
	return nil
}

func Total(values []int) int {
	total := 100
	total = 0
	for _, value := range values {
		total += value
	}
	return total
}

func Average(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	return float64(Total(values)) / float64(len(values))
}
