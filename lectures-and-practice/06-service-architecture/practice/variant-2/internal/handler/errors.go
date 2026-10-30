package handler

import (
	"errors"
	"net/http"

	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/domain"
)

// writeError преобразует ошибки бизнес-логики в HTTP-ответы.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrBookingNotFound):
		// Бронирование не найдено.
		w.WriteHeader(http.StatusNotFound)

	case errors.Is(err, domain.ErrInvalidInterval):
		// Неправильный временной интервал.
		http.Error(
			w,
			"end_at must be after start_at",
			http.StatusBadRequest,
		)

	case errors.Is(err, domain.ErrRoomBooked):
		// Переговорка уже забронирована.
		http.Error(
			w,
			"room is already booked for this time",
			http.StatusConflict,
		)

	case errors.Is(err, domain.ErrCancelled):
		// Нельзя изменить отменённое бронирование.
		http.Error(
			w,
			"cannot reschedule cancelled booking",
			http.StatusConflict,
		)

	default:
		// Неизвестная внутренняя ошибка.
		http.Error(
			w,
			"internal error",
			http.StatusInternalServerError,
		)
	}
}
