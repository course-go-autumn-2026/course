// Package main - сервис подписок и биллинга, три ручки, весь код в
// одном файле. Работает по ТЗ ниже, но архитектурно неправильно
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

// planRank - чем выше число, тем дороже тариф. Используется, чтобы
// понять, повышение это или понижение.
var planRank = map[string]int{
	"basic": 1,
	"pro":   2,
	"max":   3,
}

type storedSubscription struct {
	ID        int64
	AccountID int64
	Plan      string
	Status    string // "trial" | "active" | "suspended" | "payment_pending"
	CreatedAt time.Time
}

// storedAccount - аккаунт. Карта вместо настоящей таблицы accounts,
// но влияет на бизнес-правила подписки точно так же, как влияла бы
// запись из реальной БД.
type storedAccount struct {
	ID      int64
	HasDebt bool
}

var (
	db     = map[int64]*storedSubscription{}
	nextID int64

	// Аккаунты захардкожены прямо тут - для практики достаточно двух:
	// без задолженности и с задолженностью. account_id, которого нет
	// в карте, считается без задолженности по умолчанию.
	accounts = map[int64]*storedAccount{
		1: {ID: 1, HasDebt: false},
		2: {ID: 2, HasDebt: true},
	}

	billingURL   = "http://billing-provider.internal/charge"
	cardFraudURL = "http://card-fraud.internal/check"
)

// --- Запросы/ответы ---

type createSubscriptionRequest struct {
	AccountID int64  `json:"account_id"`
	Plan      string `json:"plan"`
}

type updatePlanRequest struct {
	Plan string `json:"plan"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /subscriptions", createSubscriptionHandler)
	mux.HandleFunc("GET /subscriptions/{id}", getSubscriptionHandler)
	mux.HandleFunc("PATCH /subscriptions/{id}/plan", updatePlanHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// createSubscriptionHandler - POST /subscriptions. Делает всё сразу:
// валидация тарифа, антифрод-проверка карты, запись в "базу", решение
// о статусе trial, формирование ответа.
func createSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Бизнес-правило №1: тариф должен быть из известного списка.
	if _, ok := planRank[req.Plan]; !ok {
		http.Error(w, "unknown plan", http.StatusBadRequest)
		return
	}

	// Проверка фрода по карте - ДО записи в базу. В отличие от
	// платёжного провайдера, недоступность антифрода не блокирует
	// оформление подписки - считаем проверку пройденной и просто
	// идём дальше (fail-open, а не fail-safe, как у провайдера).
	cfReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
		cardFraudURL, nil)
	blocked := false
	cfResp, cfErr := http.DefaultClient.Do(cfReq)
	if cfErr == nil {
		defer cfResp.Body.Close()
		if cfResp.StatusCode == http.StatusForbidden {
			blocked = true
		}
	}
	// cfErr != nil (антифрод недоступен) - blocked остаётся false, едем дальше.

	if blocked {
		http.Error(w, "card looks fraudulent", http.StatusForbidden)
		return
	}

	nextID++
	id := nextID
	sub := &storedSubscription{
		ID:        id,
		AccountID: req.AccountID,
		Plan:      req.Plan,
		Status:    "trial",
		CreatedAt: time.Now(),
	}
	db[id] = sub

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id, "status": sub.Status})
}

// getSubscriptionHandler - GET /subscriptions/{id}.
func getSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sub, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

// updatePlanHandler - PATCH /subscriptions/{id}/plan. Меняет тариф, но
// только если подписка не приостановлена - бизнес-правило №4, зашитое
// прямо тут же, рядом с разбором запроса и ответом клиенту. Дальше -
// правило №5: понижение тарифа разрешено только без задолженности у
// аккаунта, а повышение требует успешного биллинга у провайдера -
// неудача откатывает активную подписку в payment_pending.
func updatePlanHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req updatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, ok := planRank[req.Plan]; !ok {
		http.Error(w, "unknown plan", http.StatusBadRequest)
		return
	}

	sub, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Бизнес-правило №4: приостановленную подписку нельзя менять.
	if sub.Status == "suspended" {
		http.Error(w, "cannot change plan of a suspended subscription", http.StatusConflict)
		return
	}

	isUpgrade := planRank[req.Plan] > planRank[sub.Plan]
	isDowngrade := planRank[req.Plan] < planRank[sub.Plan]

	// Бизнес-правило №5a: понижение тарифа разрешено только без
	// задолженности у аккаунта.
	if isDowngrade {
		account, ok := accounts[sub.AccountID]
		if !ok {
			account = &storedAccount{ID: sub.AccountID, HasDebt: false}
		}
		if account.HasDebt {
			http.Error(w, "cannot downgrade plan with an active debt", http.StatusConflict)
			return
		}
	}

	sub.Plan = req.Plan

	// Бизнес-правило №5b: повышение тарифа требует успешного биллинга
	// у провайдера. Недоступность или отказ провайдера - fail-safe:
	// подписка откатывается в payment_pending, а не остаётся active.
	if isUpgrade && sub.Status == "active" {
		billReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
			billingURL, nil)
		billResp, billErr := http.DefaultClient.Do(billReq)
		if billErr != nil {
			sub.Status = "payment_pending"
		} else {
			defer billResp.Body.Close()
			if billResp.StatusCode != http.StatusOK {
				sub.Status = "payment_pending"
			}
			// иначе остаётся active - биллинг прошёл успешно.
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

// findByID - вспомогательная функция, имитирует поход в БД через
// database/sql, чтобы в коде было видно характерное ветвление
// sql.ErrNoRows - даже хотя реальной БД тут нет. Не используется
// напрямую в хендлерах выше (они читают db map напрямую), но
// пригодится как пример того, что нужно вынести в репозиторий.
func findByID(ctx context.Context, conn *sql.DB, id int64) (*storedSubscription, error) {
	row := conn.QueryRowContext(ctx, `SELECT * FROM subscriptions WHERE id = $1`, id)
	var s storedSubscription
	err := row.Scan(&s.ID, &s.AccountID, &s.Plan, &s.Status, &s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Ошибка "не найдено" отдаётся вызывающему коду как есть -
		// кто вызвал findByID, тот и решает, что с sql.ErrNoRows делать.
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
