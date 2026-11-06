package ready

import (
	"context"
	"net/http"
	"time"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
)

type Handler struct {
	checker  checker
	timeout  time.Duration
	response *response.Writer
}

func New(checker checker, timeout time.Duration, response *response.Writer) *Handler {
	return &Handler{checker: checker, timeout: timeout, response: response}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()
	if err := h.checker.Ping(ctx); err != nil {
		response.RecordError(ctx, err)
		h.response.JSON(w, r, http.StatusServiceUnavailable, api.HealthResponse{Status: api.Unavailable})
		return
	}
	h.response.JSON(w, r, http.StatusOK, api.HealthResponse{Status: api.Ok})
}
