package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/nestera-sh/course_06/domain"
	"github.com/nestera-sh/course_06/service"
)

type ListingHandler struct {
	svc *service.ListingService
}

func NewListingHandler(svc *service.ListingService) *ListingHandler {
	return &ListingHandler{svc: svc}
}

func (h *ListingHandler) Publish(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string  `json:"title"`
		Price    float64 `json:"price"`
		PhotoURL string  `json:"photo_url"`
		SellerID int64   `json:"seller_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	listing, err := h.svc.Create(r.Context(), service.CreateInput{
		Title:    req.Title,
		Price:    req.Price,
		PhotoURL: req.PhotoURL,
		SellerID: req.SellerID,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":     listing.ID,
		"status": listing.Status,
	})
}

func (h *ListingHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	listing, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listing)
}

func (h *ListingHandler) UpdatePrice(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	listing, err := h.svc.UpdatePrice(r.Context(), id, req.Price)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listing)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrPhotoMissing),
		errors.Is(err, domain.ErrPriceSuspicious):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, domain.ErrCannotUpdateRejected):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
