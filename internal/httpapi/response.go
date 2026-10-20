package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Smoz1e/Bulatov_Ilya_Lab/api"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/trip"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func writeProblem(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	title string,
	code string,
	detail string,
) {
	instance := r.URL.Path

	writeJSONWithContentType(
		w,
		"application/problem+json",
		status,
		api.Problem{
			Type:     "https://tripgo.example/problems/" + strings.ReplaceAll(code, "_", "-"),
			Title:    title,
			Status:   int32(status),
			Detail:   &detail,
			Instance: &instance,
			Code:     code,
		},
	)
}

func writeJSONWithContentType(
	w http.ResponseWriter,
	contentType string,
	status int,
	value any,
) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func HandleOpenAPIError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	writeProblem(
		w,
		r,
		http.StatusBadRequest,
		"Invalid request",
		"invalid_request",
		err.Error(),
	)
}

func writeServiceError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	switch {
	case errors.Is(err, trip.ErrDriverBusy):
		writeProblem(
			w,
			r,
			http.StatusConflict,
			"Driver busy",
			"driver_busy",
			"Driver already has an active trip",
		)

	case errors.Is(err, trip.ErrTripNotFound):
		writeProblem(
			w,
			r,
			http.StatusNotFound,
			"Trip not found",
			"trip_not_found",
			"Trip does not exist",
		)

	case errors.Is(err, trip.ErrTripCompleted):
		writeProblem(
			w,
			r,
			http.StatusConflict,
			"Trip completed",
			"trip_completed",
			"Trip is already completed",
		)

	default:
		writeProblem(
			w,
			r,
			http.StatusInternalServerError,
			"Internal server error",
			"internal_error",
			"An unexpected error occurred",
		)
	}
}
