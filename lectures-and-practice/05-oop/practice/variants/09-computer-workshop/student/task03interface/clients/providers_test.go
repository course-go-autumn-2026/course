package clients

import (
	"errors"
	"testing"
)

func TestSupportedOptions(t *testing.T) {
	for _, option := range []string{"Ноутбук", "Компьютер"} {
		t.Run(option, func(t *testing.T) {
			price, minutes, err := (&LocalTechnician{}).Estimate(option, 2)
			if err != nil || price != 1200 || minutes != 60 {
				t.Fatalf("price=%d, minutes=%d, err=%v", price, minutes, err)
			}
			result, err := (&PartnerTechnician{}).Calculate(ServiceRequest{
				DeviceType: option, DeviceCount: 2,
			})
			if err != nil || result.CostKopecks != 110050 || result.DurationSeconds != 2655 {
				t.Fatalf("result=%+v, err=%v", result, err)
			}
		})
	}
}

func TestUnavailableOption(t *testing.T) {
	_, _, err := (&LocalTechnician{}).Estimate("Планшет", 1)
	if !errors.Is(err, ErrLocalDeviceUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = (&PartnerTechnician{}).Calculate(ServiceRequest{
		DeviceType: "Планшет", DeviceCount: 1,
	})
	if !errors.Is(err, ErrPartnerDeviceUnavailable) {
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
		{name: "zero amount", option: "Ноутбук", amount: 0},
		{name: "negative amount", option: "Ноутбук", amount: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := (&LocalTechnician{}).Estimate(tt.option, tt.amount)
			if err == nil || errors.Is(err, ErrLocalDeviceUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
			_, err = (&PartnerTechnician{}).Calculate(ServiceRequest{
				DeviceType: tt.option, DeviceCount: tt.amount,
			})
			if err == nil || errors.Is(err, ErrPartnerDeviceUnavailable) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}
