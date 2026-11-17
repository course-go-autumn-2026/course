package domainerror

import (
	"errors"
	"net/http"
	"strings"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	api "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/generated"
)

func Problem(err error, instance string) (int, api.Problem) {
	status := http.StatusInternalServerError
	title := "Internal server error"
	code := "internal_error"
	detail := "An unexpected error occurred"

	var validationError *trip.ValidationError
	switch {
	case errors.As(err, &validationError):
		status = http.StatusBadRequest
		title = "Invalid request"
		code = "invalid_request"
		detail = validationError.Error()
	case errors.Is(err, trip.ErrTripNotFound):
		status = http.StatusNotFound
		title = "Trip not found"
		code = "trip_not_found"
		detail = "Trip does not exist"
	case errors.Is(err, trip.ErrDriverBusy):
		status = http.StatusConflict
		title = "Driver busy"
		code = "driver_busy"
		detail = "Driver already has an active trip"
	case errors.Is(err, trip.ErrTripCompleted):
		status = http.StatusConflict
		title = "Trip completed"
		code = "trip_completed"
		detail = "Trip is already completed"
	}

	return status, api.Problem{
		Type:     "https://tripgo.example/problems/" + strings.ReplaceAll(code, "_", "-"),
		Title:    title,
		Status:   int32(status),
		Detail:   &detail,
		Instance: &instance,
		Code:     code,
	}
}
