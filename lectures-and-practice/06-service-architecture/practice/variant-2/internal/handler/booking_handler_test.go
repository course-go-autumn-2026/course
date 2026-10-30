package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Проверяем обработку неправильного JSON.
func TestCreateBadJSON(t *testing.T) {
	mux := http.NewServeMux()
	h := NewBookingHandler(nil)
	RegisterRoutes(mux, h)

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings",
		strings.NewReader(`{"room_id":`),
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf(
			"Ожидали статус 400, получили %d",
			rec.Code,
		)
	}
}

// Проверяем обработку неправильного ID.
func TestGetInvalidID(t *testing.T) {
	mux := http.NewServeMux()
	h := NewBookingHandler(nil)
	RegisterRoutes(mux, h)

	req := httptest.NewRequest(
		http.MethodGet,
		"/bookings/abc",
		nil,
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf(
			"Ожидали статус 400, получили %d",
			rec.Code,
		)
	}
}
