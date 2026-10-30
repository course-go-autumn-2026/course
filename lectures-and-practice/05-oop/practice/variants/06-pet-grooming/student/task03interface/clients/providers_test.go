package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"Кошка", "Собака"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&LocalGroomer{}).Estimate(option, 5)
			if err != nil || price != 700 || minutes != 60 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&QuickGroomer{}).Calculate(CareRequest{
				PetType: option, MassKg: 5,
			})
			if err != nil || result.CostKopecks != 75050 || result.DurationSeconds != 2655 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&LocalGroomer{}).Estimate("Кролик", 1)
	if !errors.Is(err, ErrLocalPetUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&QuickGroomer{}).Calculate(CareRequest{
		PetType: "Кролик", MassKg: 1,
	})
	if !errors.Is(err, ErrQuickPetUnavailable) {
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
		{name: "zero amount", option: "Кошка", amount: 0},
		{name: "negative amount", option: "Кошка", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&LocalGroomer{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrLocalPetUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&QuickGroomer{}).Calculate(CareRequest{
				PetType: tt.option, MassKg: tt.amount,
			})
			if err == nil || errors.Is(err, ErrQuickPetUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
