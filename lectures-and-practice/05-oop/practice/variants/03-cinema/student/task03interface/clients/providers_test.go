package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"Дневной", "Вечерний"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&BoxOffice{}).Estimate(option, 2)
			if err != nil || price != 700 || minutes != 5 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&OnlineTickets{}).Calculate(TicketRequest{
				ShowType: option, SeatCount: 2,
			})
			if err != nil || result.CostKopecks != 71050 || result.DurationSeconds != 135 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&BoxOffice{}).Estimate("Ночной", 1)
	if !errors.Is(err, ErrBoxOfficeShowUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&OnlineTickets{}).Calculate(TicketRequest{
		ShowType: "Ночной", SeatCount: 1,
	})
	if !errors.Is(err, ErrOnlineShowUnavailable) {
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
		{name: "zero amount", option: "Дневной", amount: 0},
		{name: "negative amount", option: "Дневной", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&BoxOffice{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrBoxOfficeShowUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&OnlineTickets{}).Calculate(TicketRequest{
				ShowType: tt.option, SeatCount: tt.amount,
			})
			if err == nil || errors.Is(err, ErrOnlineShowUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
