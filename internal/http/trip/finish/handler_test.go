package finish_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	trip "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/domain/trip"
	"github.com/Smoz1e/Bulatov_Ilya_Lab/internal/http/trip/finish"
)

func TestHandlerFinishReturnsTrip(t *testing.T) {
	controller := gomock.NewController(t)
	usecase := NewMockUsecase(controller)
	id := uuid.New()
	usecase.EXPECT().Finish(gomock.Any(), id).Return(trip.Trip{ID: id, Status: trip.StatusCompleted, StartedAt: time.Now()}, nil)
	handler := finish.New(usecase)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+id.String()+"/finish", nil)
	recorder := httptest.NewRecorder()

	handler.FinishTrip(recorder, request, id)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestHandlerFinishMapsCompletedTrip(t *testing.T) {
	controller := gomock.NewController(t)
	usecase := NewMockUsecase(controller)
	id := uuid.New()
	usecase.EXPECT().Finish(gomock.Any(), id).Return(trip.Trip{}, trip.ErrTripCompleted)
	handler := finish.New(usecase)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+id.String()+"/finish", nil)
	recorder := httptest.NewRecorder()

	handler.FinishTrip(recorder, request, id)

	if recorder.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
}
