package create

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	api "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/generated"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/domainerror"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

type Handler struct {
	usecase usecase
}

func New(usecase usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (handler *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, _ api.CreateTripParams) {
	var request api.TripData
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		handler.writeProblem(w, r, &trip.ValidationError{Field: "body", Message: err.Error()})
		return
	}
	if err := ensureSingleJSONValue(decoder); err != nil {
		handler.writeProblem(w, r, &trip.ValidationError{Field: "body", Message: err.Error()})
		return
	}

	value, err := handler.usecase.Create(r.Context(), toInput(request))
	if err != nil {
		handler.writeProblem(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/trips/"+value.ID.String())
	writeJSON(w, http.StatusCreated, toAPITrip(value))
}

func (handler *Handler) writeProblem(w http.ResponseWriter, r *http.Request, err error) {
	status, body := domainerror.Problem(err, r.URL.Path)
	w.Header().Set("Content-Type", "application/problem+json")
	writeJSON(w, status, body)
}

func ensureSingleJSONValue(decoder *json.Decoder) error {
	var additionalValue any
	err := decoder.Decode(&additionalValue)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return err
	}

	return errors.New("request body must contain a single JSON object")
}

func toInput(request api.TripData) model.CreateTripInput {
	return model.CreateTripInput{
		UserID:   request.UserId,
		DriverID: request.DriverId,
		StartPoint: model.Coordinates{
			Latitude:  request.StartPoint.Latitude,
			Longitude: request.StartPoint.Longitude,
		},
		EndPoint: model.Coordinates{
			Latitude:  request.EndPoint.Latitude,
			Longitude: request.EndPoint.Longitude,
		},
		Price: request.Price,
	}
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
