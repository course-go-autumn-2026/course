package create

import (
	"net/http"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/domainerror"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
	tripresponse "github.com/DaleCoopTP/grishin_lab1/internal/http/trip"
)

type Handler struct {
	usecase  usecase
	response *response.Writer
}

func New(usecase usecase, response *response.Writer) *Handler {
	return &Handler{usecase: usecase, response: response}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	var request api.TripData
	if err := decodeCreateRequest(w, r, &request); err != nil {
		h.response.Error(w, r, domainerror.ErrInvalidRequest)
		return
	}
	trip, replayed, err := h.usecase.CreateWithKey(r.Context(), toInput(request), params.IdempotencyKey)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/trips/"+trip.ID.String())
	status := http.StatusCreated
	if replayed {
		status = http.StatusOK
	}
	h.response.JSON(w, r, status, tripresponse.ToResponse(trip))
}
