package httpapi

import (
	"github.com/Smoz1e/Bulatov_Ilya_Lab/api"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/trip"
)

func toAPITrip(value trip.Trip) api.Trip {
	return api.Trip{
		Id:       value.ID,
		UserId:   value.UserID,
		DriverId: value.DriverID,
		StartPoint: api.Coordinates{
			Latitude:  value.StartPoint.Latitude,
			Longitude: value.StartPoint.Longitude,
		},
		EndPoint: api.Coordinates{
			Latitude:  value.EndPoint.Latitude,
			Longitude: value.EndPoint.Longitude,
		},
		Price:          value.Price,
		Status:         api.TripStatus(value.Status),
		StartedAt:      value.StartedAt,
		FinishedAt:     value.FinishedAt,
		LastPositionAt: value.LastPositionAt,
	}
}
