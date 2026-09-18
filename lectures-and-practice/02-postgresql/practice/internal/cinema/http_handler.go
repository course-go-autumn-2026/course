package cinema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTPHandler struct {
	screenings *ScreeningRepository
	bookings   *BookingService
	logger     *log.Logger
}

func NewHTTPHandler(
	screenings *ScreeningRepository,
	bookings *BookingService,
	logger *log.Logger,
) *HTTPHandler {
	return &HTTPHandler{screenings: screenings, bookings: bookings, logger: logger}
}

func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /screenings", h.searchScreenings)
	mux.HandleFunc("POST /screenings/{screening_id}/bookings", h.createBooking)
	return mux
}

func (h *HTTPHandler) searchScreenings(w http.ResponseWriter, r *http.Request) {
	filter, err := parseScreeningFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	screenings, err := h.screenings.Search(r.Context(), filter)
	if err != nil {
		h.logger.Printf("search screenings: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, screenings)
}

func parseScreeningFilter(r *http.Request) (ScreeningFilter, error) {
	query := r.URL.Query()
	filter := ScreeningFilter{
		Search: strings.TrimSpace(query.Get("search")),
		Limit:  20,
	}

	if raw := query.Get("starts_from"); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return ScreeningFilter{}, fmt.Errorf("starts_from must use RFC3339 format")
		}
		filter.StartsFrom = &value
	}
	if raw := query.Get("min_seats"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return ScreeningFilter{}, fmt.Errorf("min_seats must be a non-negative integer")
		}
		filter.MinSeats = &value
	}
	if raw := query.Get("limit"); raw != "" {
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || value < 1 || value > 100 {
			return ScreeningFilter{}, fmt.Errorf("limit must be between 1 and 100")
		}
		filter.Limit = value
	}

	return filter, nil
}

func (h *HTTPHandler) createBooking(w http.ResponseWriter, r *http.Request) {
	screeningID, err := strconv.ParseInt(r.PathValue("screening_id"), 10, 64)
	if err != nil || screeningID < 1 {
		writeError(w, http.StatusBadRequest, "screening_id must be a positive integer")
		return
	}

	var request struct {
		UserID int64 `json:"user_id"`
	}
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request body must contain one JSON object")
		return
	}
	if request.UserID < 1 {
		writeError(w, http.StatusBadRequest, "user_id must be a positive integer")
		return
	}

	booking, err := h.bookings.Book(r.Context(), screeningID, request.UserID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoSeats), errors.Is(err, ErrAlreadyBooked):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, ErrNotImplemented):
			writeError(w, http.StatusNotImplemented, err.Error())
		default:
			h.logger.Printf("create booking: %v", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, booking)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
