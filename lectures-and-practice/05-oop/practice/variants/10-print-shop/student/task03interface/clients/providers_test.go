package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"A4", "A3"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&OfficePrinter{}).Estimate(option, 25)
			if err != nil || price != 350 || minutes != 15 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&QuickPrinter{}).Calculate(PrintRequest{
				PaperFormat: option, PageCount: 25,
			})
			if err != nil || result.CostKopecks != 35050 || result.DurationSeconds != 555 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&OfficePrinter{}).Estimate("A5", 1)
	if !errors.Is(err, ErrOfficeFormatUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&QuickPrinter{}).Calculate(PrintRequest{
		PaperFormat: "A5", PageCount: 1,
	})
	if !errors.Is(err, ErrQuickFormatUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvalidRequests(t *testing.T) {
	tests := []struct {
		name   string
		option string
		amount int
	}{
		{name: "empty option", option: "", amount: 1},
		{name: "zero amount", option: "A4", amount: 0},
		{name: "negative amount", option: "A4", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&OfficePrinter{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrOfficeFormatUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&QuickPrinter{}).Calculate(PrintRequest{
				PaperFormat: tt.option, PageCount: tt.amount,
			})
			if err == nil || errors.Is(err, ErrQuickFormatUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
