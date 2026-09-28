package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/DaleCoopTP/grishin_lab1/internal/trips"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxBodyBytes = 1 << 20

type Handler struct {
	service      *trips.Service
	pool         *pgxpool.Pool
	queryTimeout time.Duration
	logger       *slog.Logger
}

var _ api.ServerInterface = (*Handler)(nil)

func NewHandler(service *trips.Service, pool *pgxpool.Pool, queryTimeout time.Duration, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, pool: pool, queryTimeout: queryTimeout, logger: logger}
	return api.HandlerWithOptions(h, api.ChiServerOptions{
		BaseRouter: chi.NewRouter(),
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			h.writeProblem(w, r, "invalid_request")
		},
	})
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	var request api.TripData
	if err := decodeCreateRequest(w, r, &request); err != nil {
		h.writeProblem(w, r, "invalid_request")
		return
	}
	trip, replayed, err := h.service.CreateWithKey(r.Context(), trips.CreateInput{
		UserID: request.UserId, DriverID: request.DriverId,
		StartPoint: trips.Point{Latitude: request.StartPoint.Latitude, Longitude: request.StartPoint.Longitude},
		EndPoint:   trips.Point{Latitude: request.EndPoint.Latitude, Longitude: request.EndPoint.Longitude},
		Price:      request.Price,
	}, params.IdempotencyKey)
	if err != nil {
		h.writeTripError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/trips/"+trip.ID.String())
	status := http.StatusCreated
	if replayed {
		status = http.StatusOK
	}
	h.writeJSON(w, r, status, toAPITrip(trip))
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripID api.TripId) {
	trip, err := h.service.Get(r.Context(), tripID)
	if err != nil {
		h.writeTripError(w, r, err)
		return
	}
	h.writeJSON(w, r, http.StatusOK, toAPITrip(trip))
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripID api.TripId) {
	trip, err := h.service.Finish(r.Context(), tripID)
	if err != nil {
		h.writeTripError(w, r, err)
		return
	}
	h.writeJSON(w, r, http.StatusOK, toAPITrip(trip))
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, r, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.queryTimeout)
	defer cancel()
	if err := h.pool.Ping(ctx); err != nil {
		h.writeJSON(w, r, http.StatusServiceUnavailable, api.HealthResponse{Status: api.Unavailable})
		return
	}
	h.writeJSON(w, r, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func decodeCreateRequest(w http.ResponseWriter, r *http.Request, request *api.TripData) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errors.New("Content-Type must be application/json")
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	fields, err := requiredFields(body, "user_id", "driver_id", "start_point", "end_point", "price")
	if err != nil {
		return err
	}
	for _, point := range []string{"start_point", "end_point"} {
		if _, err := requiredFields(fields[point], "latitude", "longitude"); err != nil {
			return fmt.Errorf("%s: %w", point, err)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(request); err != nil {
		return fmt.Errorf("decode request body: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	if request.UserId == uuid.Nil || request.DriverId == uuid.Nil || request.Price < 0 ||
		!validPoint(request.StartPoint) || !validPoint(request.EndPoint) {
		return errors.New("request values are outside the allowed range")
	}
	return nil
}

func requiredFields(body []byte, names ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil || fields == nil {
		return nil, errors.New("expected a JSON object")
	}
	for _, name := range names {
		value, ok := fields[name]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, fmt.Errorf("%s is required", name)
		}
	}
	return fields, nil
}

func validPoint(point api.Coordinates) bool {
	return point.Latitude >= -90 && point.Latitude <= 90 && point.Longitude >= -180 && point.Longitude <= 180
}

func toAPITrip(trip trips.Trip) api.Trip {
	return api.Trip{
		Id: trip.ID, UserId: trip.UserID, DriverId: trip.DriverID,
		StartPoint: api.Coordinates{Latitude: trip.StartPoint.Latitude, Longitude: trip.StartPoint.Longitude},
		EndPoint:   api.Coordinates{Latitude: trip.EndPoint.Latitude, Longitude: trip.EndPoint.Longitude},
		Price:      trip.Price, Status: api.TripStatus(trip.Status),
		StartedAt: trip.StartedAt, FinishedAt: trip.FinishedAt,
	}
}

func (h *Handler) writeTripError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, trips.ErrNotFound):
		h.writeProblem(w, r, "trip_not_found")
	case errors.Is(err, trips.ErrCompleted):
		h.writeProblem(w, r, "trip_completed")
	case errors.Is(err, trips.ErrDriverBusy):
		h.writeProblem(w, r, "driver_busy")
	case errors.Is(err, trips.ErrIdempotencyConflict):
		h.writeProblem(w, r, "idempotency_conflict")
	default:
		h.logger.ErrorContext(r.Context(), "request failed", "error", err, "method", r.Method, "path", r.URL.Path)
		h.writeProblem(w, r, "internal_error")
	}
}

func (h *Handler) writeProblem(w http.ResponseWriter, r *http.Request, code string) {
	problem := problemFor(code, r.URL.Path)
	data, err := json.Marshal(problem)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(int(problem.Status))
	_, _ = w.Write(data)
}

func problemFor(code, path string) api.Problem {
	problem := api.Problem{Code: code, Status: http.StatusInternalServerError, Title: "Internal Server Error", Type: "https://tripgo.example/problems/internal-error"}
	detail := "Internal server error"
	switch code {
	case "invalid_request":
		problem.Status, problem.Title, problem.Type = http.StatusBadRequest, "Invalid request", "https://tripgo.example/problems/invalid-request"
		detail = "Request validation failed"
	case "trip_not_found":
		problem.Status, problem.Title, problem.Type = http.StatusNotFound, "Trip not found", "https://tripgo.example/problems/trip-not-found"
		detail = "Trip was not found"
	case "trip_completed":
		problem.Status, problem.Title, problem.Type = http.StatusConflict, "Trip completed", "https://tripgo.example/problems/trip-completed"
		detail = "Operation is not allowed for a completed trip"
	case "driver_busy":
		problem.Status, problem.Title, problem.Type = http.StatusConflict, "Driver busy", "https://tripgo.example/problems/driver-busy"
		detail = "Driver already has an active trip"
	case "idempotency_conflict":
		problem.Status, problem.Title, problem.Type = http.StatusConflict, "Idempotency conflict", "https://tripgo.example/problems/idempotency-conflict"
		detail = "Idempotency-Key was already used with a different request body"
	}
	problem.Detail = &detail
	problem.Instance = &path
	return problem
}

func (h *Handler) writeJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "encode response", "error", err)
		h.writeProblem(w, r, "internal_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
