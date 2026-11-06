package domainerror

import (
	"errors"
	"net/http"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
)

var (
	ErrInvalidRequest   = errors.New("invalid request")
	ErrRouteNotFound    = errors.New("route not found")
	ErrMethodNotAllowed = errors.New("method not allowed")
)

func Problem(err error) (int, api.Problem) {
	code := "internal_error"
	var validationError *trip.ValidationError
	switch {
	case errors.Is(err, ErrInvalidRequest), errors.As(err, &validationError):
		code = "invalid_request"
	case errors.Is(err, ErrRouteNotFound):
		code = "route_not_found"
	case errors.Is(err, ErrMethodNotAllowed):
		code = "method_not_allowed"
	case errors.Is(err, trip.ErrNotFound):
		code = "trip_not_found"
	case errors.Is(err, trip.ErrCompleted):
		code = "trip_completed"
	case errors.Is(err, trip.ErrDriverBusy):
		code = "driver_busy"
	case errors.Is(err, trip.ErrIdempotencyConflict):
		code = "idempotency_conflict"
	}
	body := problemFor(code)
	return int(body.Status), body
}

func problemFor(code string) api.Problem {
	problem := api.Problem{Code: code, Status: http.StatusInternalServerError, Title: "Internal Server Error", Type: "https://tripgo.example/problems/internal-error"}
	detail := "Internal server error"
	switch code {
	case "invalid_request":
		problem.Status, problem.Title, problem.Type = http.StatusBadRequest, "Invalid request", "https://tripgo.example/problems/invalid-request"
		detail = "Request validation failed"
	case "route_not_found":
		problem.Status, problem.Title, problem.Type = http.StatusNotFound, "Route not found", "https://tripgo.example/problems/route-not-found"
		detail = "No route matches the request path"
	case "method_not_allowed":
		problem.Status, problem.Title, problem.Type = http.StatusMethodNotAllowed, "Method not allowed", "https://tripgo.example/problems/method-not-allowed"
		detail = "The request method is not supported for this path"
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
	return problem
}
