package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/domain"
	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/service"
)

// Интерфейс для взаимодействия с сервисным слоем.
type BookingService interface {
	Create(
		ctx context.Context,
		in service.CreateBookingInput,
	) (*domain.Booking, error)

	Get(
		ctx context.Context,
		id int64,
	) (*domain.Booking, error)

	Reschedule(
		ctx context.Context,
		id int64,
		startAt, endAt time.Time,
	) (*domain.Booking, error)
}

// Структура HTTP-обработчика.
type BookingHandler struct {
	service BookingService
}

// Конструктор обработчика.
func NewBookingHandler(s BookingService) *BookingHandler {
	return &BookingHandler{
		service: s,
	}
}

// Регистрация HTTP-маршрутов.
func RegisterRoutes(mux *http.ServeMux, h *BookingHandler) {
	mux.HandleFunc("POST /bookings", h.Create)
	mux.HandleFunc("GET /bookings/{id}", h.Get)
	mux.HandleFunc("PATCH /bookings/{id}/time", h.UpdateTime)
}

// POST /bookings — создание бронирования.
func (h *BookingHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req createBookingRequest

	// Читаем JSON из тела запроса.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Преобразуем строки времени в time.Time.
	startAt, endAt, err := parseInterval(req.StartAt, req.EndAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Передаём создание бронирования сервисному слою.
	booking, err := h.service.Create(
		r.Context(),
		service.CreateBookingInput{
			RoomID:     req.RoomID,
			EmployeeID: req.EmployeeID,
			StartAt:    startAt,
			EndAt:      endAt,
		},
	)

	if err != nil {
		writeError(w, err)
		return
	}

	// Отправляем успешный ответ клиенту.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(createBookingResponse{
		ID:     booking.ID(),
		Status: string(booking.Status()),
	})
}

// GET /bookings/{id} — получение бронирования.
func (h *BookingHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Извлекаем ID из URL.
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получаем бронирование через Service.
	booking, err := h.service.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	// Возвращаем полную информацию о бронировании.
	writeBooking(w, booking)
}

// PATCH /bookings/{id}/time — изменение времени брони.
func (h *BookingHandler) UpdateTime(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Получаем ID бронирования.
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req updateTimeRequest

	// Читаем JSON с новым временем.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Преобразуем даты.
	startAt, endAt, err := parseInterval(req.StartAt, req.EndAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Вызываем сервис изменения времени.
	booking, err := h.service.Reschedule(
		r.Context(),
		id,
		startAt,
		endAt,
	)

	if err != nil {
		writeError(w, err)
		return
	}

	// Возвращаем обновлённое бронирование.
	writeBooking(w, booking)
}

// Формирование JSON-ответа с полной информацией о брони.
func writeBooking(w http.ResponseWriter, b *domain.Booking) {
	// Получаем снимок состояния бронирования из Domain.
	snapshot := b.Snapshot()

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(bookingResponse{
		ID:         snapshot.ID,
		RoomID:     snapshot.RoomID,
		EmployeeID: snapshot.EmployeeID,
		StartAt:    snapshot.StartAt,
		EndAt:      snapshot.EndAt,
		Status:     string(snapshot.Status),
		CreatedAt:  snapshot.CreatedAt,
	})
}
