package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/Smoz1e/Bulatov_Ilya_Lab/api"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/trip"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	api.Unimplemented
	pool         *pgxpool.Pool
	queryTimeout time.Duration
	service      *trip.Service
}

func NewHandler(
	pool *pgxpool.Pool,
	queryTimeout time.Duration,
	service *trip.Service,
) *Handler {
	return &Handler{
		pool:         pool,
		queryTimeout: queryTimeout,
		service:      service,
	}
}

func (handler *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, api.HealthResponse{
		Status: api.Ok,
	})
}

func (handler *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), handler.queryTimeout)
	defer cancel()

	if err := handler.pool.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, api.HealthResponse{
			Status: api.Unavailable,
		})
		return
	}

	writeJSON(w, http.StatusOK, api.HealthResponse{
		Status: api.Ok,
	})
}

func (handler *Handler) CreateTrip(
	w http.ResponseWriter,
	r *http.Request,
	_ api.CreateTripParams,
) {
	request, err := decodeCreateTrip(r)
	if err != nil {
		writeProblem(
			w,
			r,
			http.StatusBadRequest,
			"Invalid request",
			"invalid_request",
			err.Error(),
		)
		return
	}

	input, err := createTripInput(request)
	if err != nil {
		writeProblem(
			w,
			r,
			http.StatusBadRequest,
			"Invalid request",
			"invalid_request",
			err.Error(),
		)
		return
	}

	value, err := handler.service.Create(r.Context(), input)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/trips/"+value.ID.String())
	writeJSON(w, http.StatusCreated, toAPITrip(value))
}

func (handler *Handler) GetTrip(
	w http.ResponseWriter,
	r *http.Request,
	tripID api.TripId,
) {
	value, err := handler.service.GetByID(r.Context(), tripID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toAPITrip(value))
}

func (handler *Handler) FinishTrip(
	w http.ResponseWriter,
	r *http.Request,
	tripID api.TripId,
) {
	value, err := handler.service.Finish(r.Context(), tripID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toAPITrip(value))
}
