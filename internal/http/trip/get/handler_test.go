package get_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/trip/get"
)

func TestHandlerGetReturnsTrip(t *testing.T) {
	controller := gomock.NewController(t)
	usecase := NewMockUsecase(controller)
	id := uuid.New()
	usecase.EXPECT().Get(gomock.Any(), id).Return(trip.Trip{ID: id, Status: trip.StatusActive, StartedAt: time.Now()}, nil)
	handler := get.New(usecase)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil)
	recorder := httptest.NewRecorder()

	handler.GetTrip(recorder, request, id)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestHandlerGetMapsNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	usecase := NewMockUsecase(controller)
	id := uuid.New()
	usecase.EXPECT().Get(gomock.Any(), id).Return(trip.Trip{}, trip.ErrTripNotFound)
	handler := get.New(usecase)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/trips/"+id.String(), nil)
	recorder := httptest.NewRecorder()

	handler.GetTrip(recorder, request, id)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if recorder.Header().Get("Content-Type") != "application/problem+json" {
		t.Errorf("Content-Type = %q", recorder.Header().Get("Content-Type"))
	}
}
