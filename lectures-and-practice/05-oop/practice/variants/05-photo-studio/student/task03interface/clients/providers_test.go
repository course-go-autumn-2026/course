package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"Студия A", "Студия B"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&LocalStudio{}).Estimate(option, 2)
			if err != nil || price != 1800 || minutes != 20 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&PartnerStudio{}).Calculate(SessionRequest{
				RoomName: option, DurationHours: 2,
			})
			if err != nil || result.CostKopecks != 180050 || result.DurationSeconds != 855 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&LocalStudio{}).Estimate("Студия C", 1)
	if !errors.Is(err, ErrLocalRoomUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&PartnerStudio{}).Calculate(SessionRequest{
		RoomName: "Студия C", DurationHours: 1,
	})
	if !errors.Is(err, ErrPartnerRoomUnavailable) {
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
		{name: "zero amount", option: "Студия A", amount: 0},
		{name: "negative amount", option: "Студия A", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&LocalStudio{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrLocalRoomUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&PartnerStudio{}).Calculate(SessionRequest{
				RoomName: tt.option, DurationHours: tt.amount,
			})
			if err == nil || errors.Is(err, ErrPartnerRoomUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
