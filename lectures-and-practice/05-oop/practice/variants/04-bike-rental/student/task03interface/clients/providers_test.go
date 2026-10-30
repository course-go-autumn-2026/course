package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"Городской", "Горный"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&ParkRental{}).Estimate(option, 2)
			if err != nil || price != 400 || minutes != 15 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&CityRental{}).Calculate(RentalRequest{
				BikeType: option, DurationHours: 2,
			})
			if err != nil || result.CostKopecks != 39050 || result.DurationSeconds != 555 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&ParkRental{}).Estimate("Тандем", 1)
	if !errors.Is(err, ErrParkBikeUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&CityRental{}).Calculate(RentalRequest{
		BikeType: "Тандем", DurationHours: 1,
	})
	if !errors.Is(err, ErrCityBikeUnavailable) {
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
		{name: "zero amount", option: "Городской", amount: 0},
		{name: "negative amount", option: "Городской", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&ParkRental{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrParkBikeUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&CityRental{}).Calculate(RentalRequest{
				BikeType: tt.option, DurationHours: tt.amount,
			})
			if err == nil || errors.Is(err, ErrCityBikeUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
