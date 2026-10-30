package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"Американо", "Латте"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&Barista{}).Estimate(option, 3)
			if err != nil || price != 250 || minutes != 10 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&AutoBar{}).Calculate(BrewRequest{
				Recipe: option, CupCount: 3,
			})
			if err != nil || result.CostKopecks != 24050 || result.DurationSeconds != 375 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&Barista{}).Estimate("Капучино", 1)
	if !errors.Is(err, ErrBaristaDrinkUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&AutoBar{}).Calculate(BrewRequest{
		Recipe: "Капучино", CupCount: 1,
	})
	if !errors.Is(err, ErrAutoRecipeUnavailable) {
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
		{name: "zero amount", option: "Американо", amount: 0},
		{name: "negative amount", option: "Американо", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&Barista{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrBaristaDrinkUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&AutoBar{}).Calculate(BrewRequest{
				Recipe: tt.option, CupCount: tt.amount,
			})
			if err == nil || errors.Is(err, ErrAutoRecipeUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
