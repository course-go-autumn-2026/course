// Package main - сервис бронирования переговорок, три ручки, весь код
// в одном файле. Работает по ТЗ ниже, но архитектурно неправильно
// устроен - это и есть задание: привести к структуре слоёв без
// изменения поведения.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// --- Хранилище в памяти вместо настоящей БД (для практики достаточно,
// протокол обращения к нему такой же неудобный, как к *sql.DB). ---

type storedBooking struct {
	ID         int64
	RoomID     int64
	EmployeeID int64
	StartAt    time.Time
	EndAt      time.Time
	Status     string // "pending_confirmation" | "confirmed" | "cancelled" | "rescheduling"
	CreatedAt  time.Time
}

// storedEmployee - сотрудник. Карта вместо настоящей таблицы employees,
// но влияет на бизнес-правила бронирования точно так же, как влияла
// бы запись из реальной БД.
type storedEmployee struct {
	ID   int64
	Role string // "regular" | "admin"
}

var (
	db     = map[int64]*storedBooking{}
	nextID int64

	// Сотрудники захардкожены прямо тут - для практики достаточно двух:
	// обычного и админа. employee_id, которого нет в карте, считается
	// обычным ("regular") по умолчанию.
	employees = map[int64]*storedEmployee{
		1: {ID: 1, Role: "regular"},
		2: {ID: 2, Role: "admin"},
	}

	notifyURL           = "http://notify.internal/send"
	roomAvailabilityURL = "http://room-availability.internal/check"

	// Перенос брони ближе, чем за это время до начала, считается
	// "горящим" и требует повторного подтверждения.
	rescheduleConfirmWindow = 2 * time.Hour
)

// --- Запросы/ответы ---

type createBookingRequest struct {
	RoomID     int64  `json:"room_id"`
	EmployeeID int64  `json:"employee_id"`
	StartAt    string `json:"start_at"` // RFC3339
	EndAt      string `json:"end_at"`   // RFC3339
}

type updateTimeRequest struct {
	StartAt string `json:"start_at"`
	EndAt   string `json:"end_at"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /bookings", createBookingHandler)
	mux.HandleFunc("GET /bookings/{id}", getBookingHandler)
	mux.HandleFunc("PATCH /bookings/{id}/time", updateTimeHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// createBookingHandler - POST /bookings. Делает всё сразу: валидация,
// поиск пересечений по комнате, роль сотрудника, поход в сервис
// доступности комнаты, запись в "базу", рассылка уведомлений, ответ.
func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Бизнес-правило №1: интервал должен быть осмысленным.
	startAt, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		http.Error(w, "start_at must be a valid timestamp", http.StatusBadRequest)
		return
	}
	endAt, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		http.Error(w, "end_at must be a valid timestamp", http.StatusBadRequest)
		return
	}
	if !endAt.After(startAt) {
		http.Error(w, "end_at must be after start_at", http.StatusBadRequest)
		return
	}

	// Сотрудника смотрим прямо тут же, через ту же map, что и брони.
	// Неизвестный employee_id - не ошибка, просто "regular" по умолчанию.
	employee, ok := employees[req.EmployeeID]
	if !ok {
		employee = &storedEmployee{ID: req.EmployeeID, Role: "regular"}
	}

	// Бизнес-правило №2: пересечение по времени в одной комнате.
	// Обычному сотруднику пересекаться нельзя, админ пересекает чужую
	// бронь и отменяет её - правило зависит и от времени, и от роли.
	var overlapping *storedBooking
	for _, b := range db {
		if b.RoomID != req.RoomID || b.Status == "cancelled" {
			continue
		}
		if startAt.Before(b.EndAt) && b.StartAt.Before(endAt) {
			overlapping = b
			break
		}
	}
	if overlapping != nil {
		if employee.Role != "admin" {
			http.Error(w, "room is already booked for this time", http.StatusConflict)
			return
		}
		// Админ побеждает - чужая бронь отменяется прямо тут же, без
		// похода через отдельный метод модели.
		overlapping.Status = "cancelled"
	}

	// Поход в сервис доступности комнаты (обслуживание, уборка и т.п.) -
	// ДО записи в базу. Недоступность этого сервиса - fail-safe: бронь
	// создаётся, но сразу уходит в ожидание подтверждения, а не в confirmed.
	avReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
		roomAvailabilityURL, strings.NewReader(fmt.Sprintf(`{"room_id":%d,"start_at":%q,"end_at":%q}`, req.RoomID, req.StartAt, req.EndAt)))
	avReq.Header.Set("Content-Type", "application/json")
	avResp, avErr := http.DefaultClient.Do(avReq)
	roomAvailable := false
	if avErr == nil {
		defer avResp.Body.Close()
		roomAvailable = avResp.StatusCode == http.StatusOK
	}
	// avErr != nil (сервис доступности недоступен) - roomAvailable
	// остаётся false, бронь уйдёт в pending_confirmation ниже.

	nextID++
	id := nextID
	booking := &storedBooking{
		ID:         id,
		RoomID:     req.RoomID,
		EmployeeID: req.EmployeeID,
		StartAt:    startAt,
		EndAt:      endAt,
		Status:     "pending_confirmation",
		CreatedAt:  time.Now(),
	}
	db[id] = booking

	if roomAvailable {
		booking.Status = "confirmed"
	}
	// иначе остаётся "pending_confirmation" - сервис недоступен или
	// явно сказал, что комната занята обслуживанием.

	// Рассылка уведомлений участникам - в отличие от проверки
	// доступности, это НЕ блокирует создание брони и не влияет на
	// статус. Недоступность сервиса уведомлений просто молча
	// игнорируется (fail-open, а не fail-safe, как у доступности комнаты).
	notifyReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
		notifyURL, strings.NewReader(fmt.Sprintf(`{"booking_id":%d}`, id)))
	notifyReq.Header.Set("Content-Type", "application/json")
	notifyResp, notifyErr := http.DefaultClient.Do(notifyReq)
	if notifyErr == nil {
		notifyResp.Body.Close()
	}
	// notifyErr != nil - никого не уведомили, но бронь всё равно создана.

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id, "status": booking.Status})
}

// getBookingHandler - GET /bookings/{id}.
func getBookingHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	booking, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// updateTimeHandler - PATCH /bookings/{id}/time. Переносит бронь на
// новый интервал, но только если она не отменена - бизнес-правило №4,
// зашитое прямо тут же, рядом с разбором запроса и ответом клиенту.
// Дальше - правило №5: перенос ближе к началу, чем rescheduleConfirmWindow,
// переводит бронь в rescheduling - нужно подтвердить заново.
func updateTimeHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req updateTimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newStart, err := time.Parse(time.RFC3339, req.StartAt)
	if err != nil {
		http.Error(w, "start_at must be a valid timestamp", http.StatusBadRequest)
		return
	}
	newEnd, err := time.Parse(time.RFC3339, req.EndAt)
	if err != nil {
		http.Error(w, "end_at must be a valid timestamp", http.StatusBadRequest)
		return
	}
	if !newEnd.After(newStart) {
		http.Error(w, "end_at must be after start_at", http.StatusBadRequest)
		return
	}

	booking, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Бизнес-правило №4: отменённую бронь нельзя переносить.
	if booking.Status == "cancelled" {
		http.Error(w, "cannot reschedule cancelled booking", http.StatusConflict)
		return
	}

	booking.StartAt = newStart
	booking.EndAt = newEnd

	// Бизнес-правило №5: если бронь была confirmed, а новое начало
	// слишком близко (меньше rescheduleConfirmWindow до старта) -
	// нужно подтвердить заново, статус откатывается в rescheduling.
	// Для pending_confirmation/rescheduling ничего особого не
	// происходит - время просто меняется, статус остаётся как был.
	if booking.Status == "confirmed" {
		if time.Until(newStart) < rescheduleConfirmWindow {
			booking.Status = "rescheduling"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// findByID - вспомогательная функция, имитирует поход в БД через
// database/sql, чтобы в коде было видно характерное ветвление
// sql.ErrNoRows - даже хотя реальной БД тут нет. Не используется
// напрямую в хендлерах выше (они читают db map напрямую), но
// пригодится как пример того, что нужно вынести в репозиторий.
func findByID(ctx context.Context, conn *sql.DB, id int64) (*storedBooking, error) {
	row := conn.QueryRowContext(ctx, `SELECT * FROM bookings WHERE id = $1`, id)
	var b storedBooking
	err := row.Scan(&b.ID, &b.RoomID, &b.EmployeeID, &b.StartAt, &b.EndAt, &b.Status, &b.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Ошибка "не найдено" отдаётся вызывающему коду как есть -
		// кто вызвал findByID, тот и решает, что с sql.ErrNoRows делать.
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}
