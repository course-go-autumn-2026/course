package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"Маленький", "Большой"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&LocalFlorist{}).Estimate(option, 3)
			if err != nil || price != 350 || minutes != 20 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&QuickFlorist{}).Calculate(AssemblyRequest{
				BouquetSize: option, BouquetCount: 3,
			})
			if err != nil || result.CostKopecks != 37050 || result.DurationSeconds != 915 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&LocalFlorist{}).Estimate("Гигантский", 1)
	if !errors.Is(err, ErrLocalSizeUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&QuickFlorist{}).Calculate(AssemblyRequest{
		BouquetSize: "Гигантский", BouquetCount: 1,
	})
	if !errors.Is(err, ErrQuickSizeUnavailable) {
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
		{name: "zero amount", option: "Маленький", amount: 0},
		{name: "negative amount", option: "Маленький", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&LocalFlorist{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrLocalSizeUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&QuickFlorist{}).Calculate(AssemblyRequest{
				BouquetSize: tt.option, BouquetCount: tt.amount,
			})
			if err == nil || errors.Is(err, ErrQuickSizeUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
