package handler

import (
	"context"
	"net/http"

	"github.com/TheJCN/Avito-Course/api"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

func (h *Handler) Health(
	w http.ResponseWriter,
	r *http.Request,
) {
	writeJSON(w, http.StatusOK, api.HealthResponse{
		Status: api.Ok,
	})
}

func (h *Handler) Ready(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, api.HealthResponse{
			Status: api.Unavailable,
		})
		return
	}

	writeJSON(w, http.StatusOK, api.HealthResponse{
		Status: api.Ok,
	})
}
