package clients

import "errors"

var ErrParkBikeUnavailable = errors.New("ParkRental: option unavailable")
var ErrCityBikeUnavailable = errors.New("CityRental: option unavailable")

type ParkRental struct{}

func (*ParkRental) Estimate(bike string, hours int) (int, int, error) {
	if bike == "" || hours <= 0 {
		return 0, 0, errors.New("ParkRental: invalid request")
	}
	if bike != "Городской" && bike != "Горный" {
		return 0, 0, ErrParkBikeUnavailable
	}
	return 100 + 150*hours, 15, nil
}

func (*ParkRental) ListBikes() []string {
	return []string{"Первый", "Второй"}
}

type RentalRequest struct {
	BikeType      string
	DurationHours int
}

type RentalResponse struct {
	CostKopecks     int
	DurationSeconds int
}

type CityRental struct{}

func (*CityRental) Calculate(request RentalRequest) (RentalResponse, error) {
	if request.BikeType == "" || request.DurationHours <= 0 {
		return RentalResponse{}, errors.New("CityRental: invalid request")
	}
	if request.BikeType != "Городской" && request.BikeType != "Горный" {
		return RentalResponse{}, ErrCityBikeUnavailable
	}
	return RentalResponse{
		CostKopecks:     (150+120*request.DurationHours)*100 + 50,
		DurationSeconds: 555,
	}, nil
}

func (*CityRental) Unlock(id string) string {
	return "Замок открыт: " + id
}
