package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"Легковой", "Кроссовер"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&ManualWash{}).Estimate(option, 2)
			if err != nil || price != 900 || minutes != 40 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&AutoWash{}).Calculate(WashRequest{
				VehicleClass: option, VehicleCount: 2,
			})
			if err != nil || result.CostKopecks != 90050 || result.DurationSeconds != 1755 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&ManualWash{}).Estimate("Грузовой", 1)
	if !errors.Is(err, ErrManualClassUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&AutoWash{}).Calculate(WashRequest{
		VehicleClass: "Грузовой", VehicleCount: 1,
	})
	if !errors.Is(err, ErrAutoClassUnavailable) {
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
		{name: "zero amount", option: "Легковой", amount: 0},
		{name: "negative amount", option: "Легковой", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&ManualWash{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrManualClassUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&AutoWash{}).Calculate(WashRequest{
				VehicleClass: tt.option, VehicleCount: tt.amount,
			})
			if err == nil || errors.Is(err, ErrAutoClassUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
