package clients

import "errors"

var ErrLocalDeviceUnavailable = errors.New("LocalTechnician: option unavailable")
var ErrPartnerDeviceUnavailable = errors.New("PartnerTechnician: option unavailable")

type LocalTechnician struct{}

func (*LocalTechnician) Estimate(device string, count int) (int, int, error) {
	if device == "" || count <= 0 {
		return 0, 0, errors.New("LocalTechnician: invalid request")
	}
	if device != "Ноутбук" && device != "Компьютер" {
		return 0, 0, ErrLocalDeviceUnavailable
	}
	return 200 + 500*count, 60, nil
}

func (*LocalTechnician) ListTechnicians() []string {
	return []string{"Первый", "Второй"}
}

type ServiceRequest struct {
	DeviceType  string
	DeviceCount int
}

type ServiceResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type PartnerTechnician struct{}

func (*PartnerTechnician) Calculate(request ServiceRequest) (ServiceResponse, error) {
	if request.DeviceType == "" || request.DeviceCount <= 0 {
		return ServiceResponse{}, errors.New("PartnerTechnician: invalid request")
	}
	if request.DeviceType != "Ноутбук" && request.DeviceType != "Компьютер" {
		return ServiceResponse{}, ErrPartnerDeviceUnavailable
	}
	return ServiceResponse{
		CostKopecks:     (300+400*request.DeviceCount)*100 + 50,
		DurationSeconds: 2655,
	}, nil
}

func (*PartnerTechnician) Track(id string) string {
	return "Заявка выполняется: " + id
}
