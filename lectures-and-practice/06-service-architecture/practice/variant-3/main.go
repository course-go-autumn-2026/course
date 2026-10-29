// Package main - сервис заказов интернет-магазина, три ручки, весь
// код в одном файле. Работает по ТЗ ниже, но архитектурно неправильно
// устроен - это и есть задание: привести к структуре слоёв без
// изменения поведения.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

// --- Хранилище в памяти вместо настоящей БД (для практики достаточно,
// протокол обращения к нему такой же неудобный, как к *sql.DB). ---

type orderItem struct {
	SKU      string
	Quantity int
	Price    float64
}

type storedOrder struct {
	ID           int64
	CustomerID   int64
	Items        []orderItem
	Total        float64
	ConfirmedQty int    // сколько единиц товара подтвердил склад при создании
	Status       string // "created" | "confirmed" | "pending_stock" | "cancelled"
	CreatedAt    time.Time
}

// storedCustomer - клиент. Карта вместо настоящей таблицы customers,
// но влияет на бизнес-правила заказа точно так же, как влияла бы
// запись из реальной БД.
type storedCustomer struct {
	ID    int64
	Level string // "new" | "regular" | "vip"
}

var (
	db     = map[int64]*storedOrder{}
	nextID int64

	// Клиенты захардкожены прямо тут - для практики достаточно двух:
	// обычного и vip. customer_id, которого нет в карте, считается
	// новым ("new") по умолчанию.
	customers = map[int64]*storedCustomer{
		1: {ID: 1, Level: "regular"},
		2: {ID: 2, Level: "vip"},
	}

	stockCheckURL   = "http://stock-check.internal/check"
	paymentFraudURL = "http://payment-fraud.internal/check"
)

// --- Запросы/ответы ---

type createOrderRequest struct {
	CustomerID int64       `json:"customer_id"`
	Items      []orderItem `json:"items"`
}

type updateItemsRequest struct {
	Items []orderItem `json:"items"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", createOrderHandler)
	mux.HandleFunc("GET /orders/{id}", getOrderHandler)
	mux.HandleFunc("PATCH /orders/{id}/items", updateItemsHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func orderTotal(items []orderItem) float64 {
	var total float64
	for _, it := range items {
		total += it.Price * float64(it.Quantity)
	}
	return total
}

func orderQty(items []orderItem) int {
	var qty int
	for _, it := range items {
		qty += it.Quantity
	}
	return qty
}

// createOrderHandler - POST /orders. Делает всё сразу: валидация,
// уровень клиента, лимит суммы, антифрод-проверка платежа, запись в
// "базу", поход в сервис проверки склада, решение по ответу.
func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Бизнес-правило №1: заказ без позиций не имеет смысла.
	if len(req.Items) == 0 {
		http.Error(w, "order must have at least one item", http.StatusBadRequest)
		return
	}

	// Бизнес-правило №2: каждая позиция должна быть осмысленной.
	for _, it := range req.Items {
		if it.Quantity <= 0 {
			http.Error(w, "item quantity must be positive", http.StatusBadRequest)
			return
		}
		if it.Price <= 0 {
			http.Error(w, "item price must be positive", http.StatusBadRequest)
			return
		}
	}

	total := orderTotal(req.Items)

	// Клиента смотрим прямо тут же, через ту же map, что и заказы.
	// Неизвестный customer_id - не ошибка, просто "new" по умолчанию.
	customer, ok := customers[req.CustomerID]
	if !ok {
		customer = &storedCustomer{ID: req.CustomerID, Level: "new"}
	}

	// Бизнес-правило №3: лимит суммы заказа без доп. проверки зависит
	// от уровня клиента - у vip лимит выше. Проверка суммы не только
	// про саму сумму, но и про контекст клиента.
	orderLimit := 100_000.0
	if customer.Level == "vip" {
		orderLimit = 1_000_000.0
	}
	if total > orderLimit {
		http.Error(w, "order total exceeds limit for this customer", http.StatusBadRequest)
		return
	}

	// Антифрод-проверка платежа - ДО проверки склада и ДО записи в
	// базу. В отличие от проверки склада, недоступность антифрода не
	// блокирует заказ и не переводит его в особый статус - считаем
	// проверку пройденной и просто идём дальше (fail-open, а не
	// fail-safe, как у склада).
	afReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
		paymentFraudURL, nil)
	blocked := false
	afResp, afErr := http.DefaultClient.Do(afReq)
	if afErr == nil {
		defer afResp.Body.Close()
		if afResp.StatusCode == http.StatusForbidden {
			blocked = true
		}
	}
	// afErr != nil (антифрод недоступен) - blocked остаётся false, едем дальше.

	if blocked {
		http.Error(w, "payment looks fraudulent", http.StatusForbidden)
		return
	}

	nextID++
	id := nextID
	order := &storedOrder{
		ID:         id,
		CustomerID: req.CustomerID,
		Items:      req.Items,
		Total:      total,
		Status:     "created",
		CreatedAt:  time.Now(),
	}
	db[id] = order

	// Поход во внешний сервис проверки склада прямо из хендлера.
	// Недоступность или отказ склада - fail-safe: заказ уходит в
	// pending_stock, а не в confirmed.
	stReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
		stockCheckURL, nil)
	stResp, stErr := http.DefaultClient.Do(stReq)
	if stErr != nil {
		// Склад недоступен - заказ уже создан, остаётся pending_stock.
		order.Status = "pending_stock"
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": id, "status": order.Status})
		return
	}
	defer stResp.Body.Close()

	if stResp.StatusCode == http.StatusOK {
		order.Status = "confirmed"
		order.ConfirmedQty = orderQty(req.Items)
	} else {
		order.Status = "pending_stock"
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id, "status": order.Status})
}

// getOrderHandler - GET /orders/{id}.
func getOrderHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// updateItemsHandler - PATCH /orders/{id}/items. Меняет состав заказа,
// но только если заказ ещё не отменён - бизнес-правило №4, зашитое
// прямо тут же, рядом с разбором запроса и ответом клиенту. Дальше -
// правило №5: если новый состав требует больше товара, чем подтвердил
// склад при создании, заказ откатывается в pending_stock.
func updateItemsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req updateItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "order must have at least one item", http.StatusBadRequest)
		return
	}

	order, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Бизнес-правило №4: отменённый заказ нельзя редактировать.
	if order.Status == "cancelled" {
		http.Error(w, "cannot update cancelled order", http.StatusConflict)
		return
	}

	order.Items = req.Items
	order.Total = orderTotal(req.Items)

	// Бизнес-правило №5: если заказ был confirmed, а новый состав
	// требует больше товара, чем подтвердил склад при создании -
	// нужна повторная проверка склада, статус откатывается в
	// pending_stock. Для created/pending_stock ничего особого не
	// происходит - состав просто меняется, статус остаётся как был.
	if order.Status == "confirmed" {
		if orderQty(req.Items) > order.ConfirmedQty {
			order.Status = "pending_stock"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// findByID - вспомогательная функция, имитирует поход в БД через
// database/sql, чтобы в коде было видно характерное ветвление
// sql.ErrNoRows - даже хотя реальной БД тут нет. Не используется
// напрямую в хендлерах выше (они читают db map напрямую), но
// пригодится как пример того, что нужно вынести в репозиторий.
func findByID(ctx context.Context, conn *sql.DB, id int64) (*storedOrder, error) {
	row := conn.QueryRowContext(ctx, `SELECT * FROM orders WHERE id = $1`, id)
	var o storedOrder
	err := row.Scan(&o.ID, &o.CustomerID, &o.Total, &o.Status, &o.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Ошибка "не найдено" отдаётся вызывающему коду как есть -
		// кто вызвал findByID, тот и решает, что с sql.ErrNoRows делать.
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}
