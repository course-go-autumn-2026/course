package handler

import (
	"errors"
	"time"
)

// Данные для создания бронирования
type createBookingRequest struct {
	RoomID     int64  `json:"room_id"`
	EmployeeID int64  `json:"employee_id"`
	StartAt    string `json:"start_at"`
	EndAt      string `json:"end_at"`
}

// Данные для изменения времени бронирования
type updateTimeRequest struct {
	StartAt string `json:"start_at"`
	EndAt   string `json:"end_at"`
}

// Ответ после создания бронирования
type createBookingResponse struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// Полная информация о бронировании
type bookingResponse struct {
	ID         int64
	RoomID     int64
	EmployeeID int64
	StartAt    time.Time
	EndAt      time.Time
	Status     string
	CreatedAt  time.Time
}

// Проверка формата даты и времени
func parseInterval(start, end string) (time.Time, time.Time, error) {
	startAt, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("start_at must be a valid timestamp")
	}

	endAt, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("end_at must be a valid timestamp")
	}

	return startAt, endAt, nil
}
