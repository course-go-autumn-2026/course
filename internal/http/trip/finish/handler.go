package finish

import (
	"encoding/json"
	"net/http"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	api "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/generated"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/domainerror"
)

type Handler struct {
	usecase usecase
}

func New(usecase usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (handler *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	value, err := handler.usecase.Finish(r.Context(), id)
	if err != nil {
		handler.writeProblem(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toAPITrip(value))
}

func (handler *Handler) writeProblem(w http.ResponseWriter, r *http.Request, err error) {
	status, body := domainerror.Problem(err, r.URL.Path)
	w.Header().Set("Content-Type", "application/problem+json")
	writeJSON(w, status, body)
}

func toAPITrip(value trip.Trip) api.Trip {
	return api.Trip{
		DriverId: value.DriverID,
		EndPoint: api.Coordinates{
			Latitude:  value.EndPoint.Latitude,
			Longitude: value.EndPoint.Longitude,
		},
		FinishedAt:     value.FinishedAt,
		Id:             value.ID,
		LastPositionAt: value.LastPositionAt,
		Price:          value.Price,
		StartPoint: api.Coordinates{
			Latitude:  value.StartPoint.Latitude,
			Longitude: value.StartPoint.Longitude,
		},
		StartedAt: value.StartedAt,
		Status:    api.TripStatus(value.Status),
		UserId:    value.UserID,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}
