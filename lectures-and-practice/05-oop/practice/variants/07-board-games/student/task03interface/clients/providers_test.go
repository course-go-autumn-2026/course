package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"Семейная", "Стратегия"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&ClubLibrary{}).Estimate(option, 3)
			if err != nil || price != 400 || minutes != 10 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&PartnerLibrary{}).Calculate(RentalRequest{
				GameCategory: option, RentalDays: 3,
			})
			if err != nil || result.CostKopecks != 39050 || result.DurationSeconds != 255 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&ClubLibrary{}).Estimate("Коллекционная", 1)
	if !errors.Is(err, ErrClubCategoryUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&PartnerLibrary{}).Calculate(RentalRequest{
		GameCategory: "Коллекционная", RentalDays: 1,
	})
	if !errors.Is(err, ErrPartnerCategoryUnavailable) {
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
		{name: "zero amount", option: "Семейная", amount: 0},
		{name: "negative amount", option: "Семейная", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&ClubLibrary{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrClubCategoryUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&PartnerLibrary{}).Calculate(RentalRequest{
				GameCategory: tt.option, RentalDays: tt.amount,
			})
			if err == nil || errors.Is(err, ErrPartnerCategoryUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
