package finish

import (
	"net/http"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
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

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request, id api.TripId) {
	trip, err := h.usecase.Finish(r.Context(), id)
	if err != nil {
		h.response.Error(w, r, err)
		return
	}
	h.response.JSON(w, r, http.StatusOK, tripresponse.ToResponse(trip))
}
