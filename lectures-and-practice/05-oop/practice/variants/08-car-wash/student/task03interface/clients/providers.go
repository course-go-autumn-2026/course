package clients

import "errors"

var ErrManualClassUnavailable = errors.New("ManualWash: option unavailable")
var ErrAutoClassUnavailable = errors.New("AutoWash: option unavailable")

type ManualWash struct{}

func (*ManualWash) Estimate(carClass string, cars int) (int, int, error) {
	if carClass == "" || cars <= 0 {
		return 0, 0, errors.New("ManualWash: invalid request")
	}
	if carClass != "Легковой" && carClass != "Кроссовер" {
		return 0, 0, ErrManualClassUnavailable
	}
	return 100 + 400*cars, 40, nil
}

func (*ManualWash) ListWorkers() []string {
	return []string{"Первый", "Второй"}
}

type WashRequest struct {
	VehicleClass string
	VehicleCount int
}

type WashResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type AutoWash struct{}

func (*AutoWash) Calculate(request WashRequest) (WashResponse, error) {
	if request.VehicleClass == "" || request.VehicleCount <= 0 {
		return WashResponse{}, errors.New("AutoWash: invalid request")
	}
	if request.VehicleClass != "Легковой" && request.VehicleClass != "Кроссовер" {
		return WashResponse{}, ErrAutoClassUnavailable
	}
	return WashResponse{
		CostKopecks:     (200+350*request.VehicleCount)*100 + 50,
		DurationSeconds: 1755,
	}, nil
}

func (*AutoWash) Start(id string) string {
	return "Мойка запущена: " + id
}
