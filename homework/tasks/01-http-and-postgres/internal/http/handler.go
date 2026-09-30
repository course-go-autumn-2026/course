package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whxtelyy/course/internal/domain"
	api "github.com/whxtelyy/course/internal/generated"
	"github.com/whxtelyy/course/internal/service"
)

type Handler struct {
	service      *service.TripService
	db           *pgxpool.Pool
	logger       Logger
	readyTimeout time.Duration
}

type Logger interface {
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

func NewHandler(service *service.TripService, db *pgxpool.Pool, logger Logger, readyTimeout time.Duration) *Handler {
	return &Handler{service: service, db: db, logger: logger, readyTimeout: readyTimeout}
}

var _ api.ServerInterface = (*Handler)(nil)

func (h *Handler) GeneratedErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	writeProblem(w, r, http.StatusBadRequest, "invalid_request", "Invalid request", "Request parameter is invalid", "https://tripgo.example/problems/invalid-request")
	h.logger.Error("invalid request parameter", "error", err.Error())
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	request, err := decodeTripData(r)
	if err != nil {
		writeProblem(
			w,
			r,
			http.StatusBadRequest,
			"invalid_request",
			"Invalid request",
			err.Error(),
			"https://tripgo.example/problems/invalid-request",
		)
		return
	}

	var idempotencyKey *uuid.UUID
	var requestHash string

	if params.IdempotencyKey != nil {
		key := uuid.UUID(*params.IdempotencyKey)
		idempotencyKey = &key

		requestHash, err = hashTripData(request)
		if err != nil {
			h.logger.Error("failed to hash request", "error", err.Error())
			writeProblem(
				w,
				r,
				http.StatusInternalServerError,
				"internal_error",
				"Internal Server Error",
				"Internal server error",
				"https://tripgo.example/problems/internal-error",
			)
			return
		}
	}

	result, err := h.service.Create(r.Context(), service.CreateTripRequest{
		UserID:   request.UserId,
		DriverID: request.DriverId,
		StartPoint: domain.Coordinates{
			Latitude:  request.StartPoint.Latitude,
			Longitude: request.StartPoint.Longitude,
		},
		EndPoint: domain.Coordinates{
			Latitude:  request.EndPoint.Latitude,
			Longitude: request.EndPoint.Longitude,
		},
		Price:          request.Price,
		IdempotencyKey: idempotencyKey,
		RequestHash:    requestHash,
	})
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	location := "/api/v1/trips/" + result.Trip.ID.String()
	w.Header().Set("Location", location)

	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}

	writeJSON(w, status, toAPITrip(result.Trip))

	if result.Replayed {
		h.logger.Info("idempotent trip replay", "trip_id", result.Trip.ID.String())
	} else {
		h.logger.Info("trip created", "trip_id", result.Trip.ID.String())
	}
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripID uuid.UUID) {
	trip, err := h.service.Get(r.Context(), tripID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPITrip(trip))
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripID uuid.UUID) {
	trip, err := h.service.Finish(r.Context(), tripID)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPITrip(trip))
	h.logger.Info("trip finished", "trip_id", trip.ID.String())
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.HealthResponse{Status: api.HealthResponseStatus("ok")})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.readyTimeout)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, api.HealthResponse{Status: api.HealthResponseStatus("unavailable")})
		return
	}
	writeJSON(w, http.StatusOK, api.HealthResponse{Status: api.HealthResponseStatus("ok")})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrTripNotFound):
		writeProblem(w, r, http.StatusNotFound, "trip_not_found", "Trip not found", "Trip was not found", "https://tripgo.example/problems/trip-not-found")
	case errors.Is(err, domain.ErrTripCompleted):
		writeProblem(w, r, http.StatusConflict, "trip_completed", "Trip completed", "Operation is not allowed for a completed trip", "https://tripgo.example/problems/trip-completed")
	case errors.Is(err, domain.ErrDriverBusy):
		writeProblem(w, r, http.StatusConflict, "driver_busy", "Driver busy", "Driver already has an active trip", "https://tripgo.example/problems/driver-busy")
	case errors.Is(err, domain.ErrIdempotencyConflict):
		writeProblem(w, r, http.StatusConflict, "idempotency_conflict", "Idempotency Conflict", "Idempotency-Key was already used with a different request body", "https://tripgo.example/problems/idempotency-conflict")
	default:
		h.logger.Error("request failed", "error", err.Error())
		writeProblem(w, r, http.StatusInternalServerError, "internal_error", "Internal Server Error", "Internal server error", "https://tripgo.example/problems/internal-error")
	}
}

func decodeTripData(r *http.Request) (api.TripData, error) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&raw); err != nil {
		return api.TripData{}, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return api.TripData{}, err
	}
	if raw == nil {
		return api.TripData{}, errors.New("request body must be a JSON object")
	}

	allowed := map[string]struct{}{
		"user_id": {}, "driver_id": {}, "start_point": {}, "end_point": {}, "price": {},
	}
	for field := range raw {
		if _, ok := allowed[field]; !ok {
			return api.TripData{}, fmt.Errorf("unknown field %q", field)
		}
	}
	for _, field := range []string{"user_id", "driver_id", "start_point", "end_point", "price"} {
		if _, ok := raw[field]; !ok {
			return api.TripData{}, fmt.Errorf("field %q is required", field)
		}
	}

	if err := validateCoordinateJSON("start_point", raw["start_point"]); err != nil {
		return api.TripData{}, err
	}
	if err := validateCoordinateJSON("end_point", raw["end_point"]); err != nil {
		return api.TripData{}, err
	}
	if string(raw["price"]) == "null" {
		return api.TripData{}, errors.New("field \"price\" cannot be null")
	}

	var data api.TripData
	body, err := json.Marshal(raw)
	if err != nil {
		return api.TripData{}, fmt.Errorf("marshal request: %w", err)
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return api.TripData{}, fmt.Errorf("decode request: %w", err)
	}
	if data.UserId == uuid.Nil || data.DriverId == uuid.Nil {
		return api.TripData{}, errors.New("user_id and driver_id must be non-empty UUIDs")
	}
	if err := validateCoordinates("start_point", data.StartPoint); err != nil {
		return api.TripData{}, err
	}
	if err := validateCoordinates("end_point", data.EndPoint); err != nil {
		return api.TripData{}, err
	}
	if data.Price < 0 {
		return api.TripData{}, errors.New("price must be greater than or equal to zero")
	}
	return data, nil
}

func validateCoordinateJSON(name string, value json.RawMessage) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(value, &raw); err != nil || raw == nil {
		return fmt.Errorf("%s must be a JSON object", name)
	}
	allowed := map[string]struct{}{"latitude": {}, "longitude": {}}
	for field := range raw {
		if _, ok := allowed[field]; !ok {
			return fmt.Errorf("unknown field %q in %s", field, name)
		}
	}
	for _, field := range []string{"latitude", "longitude"} {
		value, ok := raw[field]
		if !ok {
			return fmt.Errorf("field %q in %s is required", field, name)
		}
		if string(value) == "null" {
			return fmt.Errorf("field %q in %s cannot be null", field, name)
		}
		var number float64
		if err := json.Unmarshal(value, &number); err != nil {
			return fmt.Errorf("field %q in %s must be a number", field, name)
		}
	}
	return nil
}

func validateCoordinates(name string, point api.Coordinates) error {
	if point.Latitude < -90 || point.Latitude > 90 {
		return fmt.Errorf("%s.latitude must be between -90 and 90", name)
	}
	if point.Longitude < -180 || point.Longitude > 180 {
		return fmt.Errorf("%s.longitude must be between -180 and 180", name)
	}
	return nil
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain exactly one JSON value")
		}
		return fmt.Errorf("invalid trailing JSON: %w", err)
	}
	return nil
}

func toAPITrip(trip domain.Trip) api.Trip {
	return api.Trip{
		Id:             trip.ID,
		UserId:         trip.UserID,
		DriverId:       trip.DriverID,
		StartPoint:     api.Coordinates{Latitude: trip.StartPoint.Latitude, Longitude: trip.StartPoint.Longitude},
		EndPoint:       api.Coordinates{Latitude: trip.EndPoint.Latitude, Longitude: trip.EndPoint.Longitude},
		Price:          trip.Price,
		Status:         api.TripStatus(trip.Status),
		StartedAt:      trip.StartedAt,
		FinishedAt:     trip.FinishedAt,
		LastPositionAt: trip.LastPositionAt,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, title, detail, problemType string) {
	instance := r.URL.Path
	problem := api.Problem{
		Type:     problemType,
		Title:    title,
		Status:   int32(status),
		Detail:   &detail,
		Instance: &instance,
		Code:     code,
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem)
}

func hashTripData(data api.TripData) (string, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal trip data for idempotency: %w", err)
	}

	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}
