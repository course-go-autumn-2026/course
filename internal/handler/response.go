package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/TheJCN/Avito-Course/api"
	"github.com/TheJCN/Avito-Course/internal/model"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("write response", "error", err)
	}
}

func writeProblem(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	code, name, title, detail string,
) {
	path := r.URL.Path

	problem := api.Problem{
		Type:     "https://tripgo.example/problems/" + name,
		Title:    title,
		Status:   int32(status),
		Code:     code,
		Detail:   &detail,
		Instance: &path,
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(problem); err != nil {
		slog.Error("write problem", "error", err)
	}
}

func badRequest(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	slog.Debug("invalid request", "error", err)

	writeProblem(
		w, r,
		http.StatusBadRequest,
		"invalid_request",
		"invalid-request",
		"Invalid request",
		"Request validation failed",
	)
}

func writeError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	switch {
	case errors.Is(err, model.ErrTripNotFound):
		writeProblem(
			w, r,
			http.StatusNotFound,
			"trip_not_found",
			"trip-not-found",
			"Trip not found",
			"Trip was not found",
		)

	case errors.Is(err, model.ErrDriverBusy):
		writeProblem(
			w, r,
			http.StatusConflict,
			"driver_busy",
			"driver-busy",
			"Driver busy",
			"Driver already has an active trip",
		)

	case errors.Is(err, model.ErrTripCompleted):
		writeProblem(
			w, r,
			http.StatusConflict,
			"trip_completed",
			"trip-completed",
			"Trip completed",
			"Trip is already completed",
		)

	default:
		slog.Error("request failed", "error", err)

		writeProblem(
			w, r,
			http.StatusInternalServerError,
			"internal_error",
			"internal-error",
			"Internal Server Error",
			"Internal server error",
		)
	}
}
