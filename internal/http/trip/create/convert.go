package create

import (
	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/DaleCoopTP/grishin_lab1/internal/model"
)

func toInput(request api.TripData) model.CreateTripInput {
	return model.CreateTripInput{
		UserID: request.UserId, DriverID: request.DriverId,
		StartPoint: trip.Point{Latitude: request.StartPoint.Latitude, Longitude: request.StartPoint.Longitude},
		EndPoint:   trip.Point{Latitude: request.EndPoint.Latitude, Longitude: request.EndPoint.Longitude},
		Price:      request.Price,
	}
}
