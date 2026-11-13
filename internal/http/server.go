package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	api "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/generated"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/domainerror"
	createhandler "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/trip/create"
	finishhandler "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/trip/finish"
	gethandler "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/trip/get"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/model"
)

type pinger interface {
	Ping(context.Context) error
}

type createTripUsecase interface {
	Create(context.Context, model.CreateTripInput) (trip.Trip, error)
}

type getTripUsecase interface {
	Get(context.Context, api.TripId) (trip.Trip, error)
}

type finishTripUsecase interface {
	Finish(context.Context, api.TripId) (trip.Trip, error)
}

type Server struct {
	api.Unimplemented
	create       *createhandler.Handler
	get          *gethandler.Handler
	finish       *finishhandler.Handler
	pinger       pinger
	queryTimeout time.Duration
}

func NewServer(
	create createTripUsecase,
	get getTripUsecase,
	finish finishTripUsecase,
	pinger pinger,
	queryTimeout time.Duration,
) *Server {
	return &Server{
		create:       createhandler.New(create),
		get:          gethandler.New(get),
		finish:       finishhandler.New(finish),
		pinger:       pinger,
		queryTimeout: queryTimeout,
	}
}

func (server *Server) CreateTrip(w http.ResponseWriter, r *http.Request, parameters api.CreateTripParams) {
	server.create.CreateTrip(w, r, parameters)
}

func (server *Server) GetTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	server.get.GetTrip(w, r, id)
}

func (server *Server) FinishTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	server.finish.FinishTrip(w, r, id)
}

func (server *Server) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func (server *Server) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), server.queryTimeout)
	defer cancel()

	if err := server.pinger.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, api.HealthResponse{Status: api.Unavailable})
		return
	}

	writeJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func HandleOpenAPIError(w http.ResponseWriter, r *http.Request, err error) {
	WriteProblem(w, r, &trip.ValidationError{Field: "request", Message: err.Error()})
}

func WriteProblem(w http.ResponseWriter, r *http.Request, err error) {
	status, body := domainerror.Problem(err, r.URL.Path)
	w.Header().Set("Content-Type", "application/problem+json")
	writeJSON(w, status, body)
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
