package create_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	api "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/generated"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/trip/create"
)

func TestHandlerCreateReturnsCreatedTrip(t *testing.T) {
	controller := gomock.NewController(t)
	usecase := NewMockUsecase(controller)
	id := uuid.New()
	usecase.EXPECT().Create(gomock.Any(), gomock.Any()).Return(trip.Trip{ID: id, Status: trip.StatusActive, StartedAt: time.Now()}, nil)
	handler := create.New(usecase)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{"user_id":"11111111-1111-4111-8111-111111111111","driver_id":"22222222-2222-4222-8222-222222222222","start_point":{"latitude":59.9,"longitude":30.3},"end_point":{"latitude":59.8,"longitude":30.2},"price":1450}`))
	recorder := httptest.NewRecorder()
	handler.CreateTrip(recorder, request, api.CreateTripParams{})

	if recorder.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Header().Get("Location") != "/api/v1/trips/"+id.String() {
		t.Errorf("Location = %q", recorder.Header().Get("Location"))
	}
}

func TestHandlerCreateReturnsProblemForMalformedJSON(t *testing.T) {
	controller := gomock.NewController(t)
	handler := create.New(NewMockUsecase(controller))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{"user_id":`))
	recorder := httptest.NewRecorder()

	handler.CreateTrip(recorder, request, api.CreateTripParams{})

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if recorder.Header().Get("Content-Type") != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", recorder.Header().Get("Content-Type"))
	}
}

func TestHandlerCreateReturnsProblemForSecondJSONDocument(t *testing.T) {
	controller := gomock.NewController(t)
	handler := create.New(NewMockUsecase(controller))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(`{} {}`))
	recorder := httptest.NewRecorder()

	handler.CreateTrip(recorder, request, api.CreateTripParams{})

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
