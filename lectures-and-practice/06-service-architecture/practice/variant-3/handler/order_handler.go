package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/course-go-autumn-2026/practice6-variant3-order/domain"
	"github.com/course-go-autumn-2026/practice6-variant3-order/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

type OrderItemRequest struct {
	SKU      string  `json:"sku"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type CreateOrderRequest struct {
	CustomerID int64              `json:"customer_id"`
	Items      []OrderItemRequest `json:"items"`
}

type UpdateItemsRequest struct {
	Items []OrderItemRequest `json:"items"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	order, err := h.service.Create(r.Context(), service.CreateOrderInput{
		CustomerID: req.CustomerID, Items: requestItems(req.Items),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": order.ID, "status": order.Status})
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	order, err := h.service.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) UpdateItems(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req UpdateItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	order, err := h.service.UpdateItems(r.Context(), id, requestItems(req.Items))
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func requestItems(items []OrderItemRequest) []domain.Item {
	result := make([]domain.Item, len(items))
	for i, item := range items {
		result[i] = domain.Item{SKU: item.SKU, Quantity: item.Quantity, Price: item.Price}
	}
	return result
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrItemsMissing),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrOrderLimit):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrPaymentFraud):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, domain.ErrOrderCancelled):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
