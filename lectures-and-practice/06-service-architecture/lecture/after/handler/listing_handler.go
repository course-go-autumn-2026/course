// Package handler - транспортный слой. Знает про net/http
// и JSON, не знает про SQL и ничего не знает про то, как устроена
// модерация. Сравните размер и содержание этого файла
// с before/handler.go.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/course-go-autumn-2026/lecture6-architecture/after/domain"
	"github.com/course-go-autumn-2026/lecture6-architecture/after/service"
)

type ListingHandler struct {
	service *service.ListingService
}

func NewListingHandler(s *service.ListingService) *ListingHandler {
	return &ListingHandler{service: s}
}

type publishRequest struct {
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	PhotoURL string  `json:"photo_url"`
	SellerID int64   `json:"seller_id"`
}

func (h *ListingHandler) Publish(w http.ResponseWriter, r *http.Request) {
	var req publishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	listing, err := h.service.Publish(r.Context(), service.PublishInput{
		Title:    req.Title,
		Price:    req.Price,
		PhotoURL: req.PhotoURL,
		SellerID: req.SellerID,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(listing)
}

func (h *ListingHandler) Get(w http.ResponseWriter, r *http.Request, id int64) {
	listing, err := h.service.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listing)
}

// writeError - единственное место транспортного слоя, которое
// решает, какой HTTP-статус соответствует доменной ошибке.
// Ошибка не протекает наверх в сыром виде - её явно классифицируют.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidPrice), errors.Is(err, domain.ErrPhotoMissing):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
