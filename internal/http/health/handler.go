package health

import (
	"net/http"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
)

type Handler struct {
	response *response.Writer
}

func New(response *response.Writer) *Handler {
	return &Handler{response: response}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.response.JSON(w, r, http.StatusOK, api.HealthResponse{Status: api.Ok})
}
