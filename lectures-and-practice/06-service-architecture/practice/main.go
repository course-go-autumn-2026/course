// Package main - сервис объявлений, три ручки, весь код в одном файле.
// Работает по ТЗ ниже, но архитектурно неправильно устроен - это и
// есть задание: привести к структуре слоёв без изменения поведения.
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

type storedListing struct {
	ID        int64
	Title     string
	Price     float64
	PhotoURL  string
	SellerID  int64
	Status    string // "pending" | "approved" | "rejected" | "review"
	CreatedAt time.Time
}

// storedSeller - продавец. Карта вместо настоящей таблицы sellers,
// но влияет на бизнес-правила объявления точно так же, как влияла
// бы запись из реальной БД.
type storedSeller struct {
	ID   int64
	Tier string // "regular" | "trusted"
}

var (
	db     = map[int64]*storedListing{}
	nextID int64

	// Продавцы захардкожены прямо тут - для практики достаточно двух:
	// обычного и доверенного. seller_id, которого нет в карте, считается
	// обычным ("regular") по умолчанию.
	sellers = map[int64]*storedSeller{
		1: {ID: 1, Tier: "regular"},
		2: {ID: 2, Tier: "trusted"},
	}

	moderationURL = "http://moderation.internal/check"
	antifraudURL  = "http://antifraud.internal/check"
)

// --- Запросы/ответы ---

type createListingRequest struct {
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	PhotoURL string  `json:"photo_url"`
	SellerID int64   `json:"seller_id"`
}

type updatePriceRequest struct {
	Price float64 `json:"price"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /listings", createListingHandler)
	mux.HandleFunc("GET /listings/{id}", getListingHandler)
	mux.HandleFunc("PATCH /listings/{id}/price", updatePriceHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// createListingHandler - POST /listings. Делает всё сразу: валидация,
// поиск продавца, поход в антифрод, порог цены по его tier, запись в
// "базу", поход в модерацию, решение по ответу.
func createListingHandler(w http.ResponseWriter, r *http.Request) {
	var req createListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Бизнес-правило №1: нет цены - нет публикации.
	if req.Price <= 0 {
		http.Error(w, "price must be positive", http.StatusBadRequest)
		return
	}

	// Бизнес-правило №2: нет фото - нет публикации.
	if strings.TrimSpace(req.PhotoURL) == "" {
		http.Error(w, "photo is required", http.StatusBadRequest)
		return
	}

	// Продавца смотрим прямо тут же, через ту же map, что и объявления.
	// Неизвестный seller_id - не ошибка, просто "regular" по умолчанию.
	seller, ok := sellers[req.SellerID]
	if !ok {
		seller = &storedSeller{ID: req.SellerID, Tier: "regular"}
	}

	// Бизнес-правило №3: порог подозрительной цены зависит от того, кто
	// продавец - у доверенных продавцов лимит выше. Проверка цены не
	// только про саму цену, но и про контекст продавца.
	priceLimit := 10_000_000.0
	if seller.Tier == "trusted" {
		priceLimit = 50_000_000.0
	}
	if req.Price > priceLimit {
		http.Error(w, "price looks suspicious", http.StatusBadRequest)
		return
	}

	// Поход в антифрод - ДО модерации и ДО записи в базу. В отличие от
	// модерации, недоступность антифрода не блокирует публикацию и не
	// переводит объявление в особый статус - считаем проверку пройденной
	// и просто идём дальше (fail-open, а не fail-safe, как у модерации).
	afReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
		antifraudURL, strings.NewReader(fmt.Sprintf(`{"seller_id":%d,"price":%f}`, req.SellerID, req.Price)))
	afReq.Header.Set("Content-Type", "application/json")
	blocked := false
	afResp, afErr := http.DefaultClient.Do(afReq)
	if afErr == nil {
		defer afResp.Body.Close()
		if afResp.StatusCode == http.StatusForbidden {
			blocked = true
		}
	}
	// afErr != nil (антифрод недоступен) - blocked остаётся false, едем дальше.

	nextID++
	id := nextID
	listing := &storedListing{
		ID:        id,
		Title:     req.Title,
		Price:     req.Price,
		PhotoURL:  req.PhotoURL,
		SellerID:  req.SellerID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	db[id] = listing

	// Антифрод заблокировал - сразу rejected, в модерацию не идём.
	if blocked {
		listing.Status = "rejected"
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": id, "status": listing.Status})
		return
	}

	// Доверенный продавец минует модерацию - сразу approved, запрос в
	// moderationURL вообще не делаем.
	if seller.Tier == "trusted" {
		listing.Status = "approved"
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": id, "status": listing.Status})
		return
	}

	// Поход во внешний сервис модерации прямо из хендлера. Если тело
	// запроса содержит "slow" - fake-сервис в тестах держит соединение
	// дольше контекстного таймаута, это и есть эмуляция "не ответила
	// вовремя" для статуса review.
	modCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	modReq, _ := http.NewRequestWithContext(modCtx, http.MethodPost,
		moderationURL, strings.NewReader(fmt.Sprintf(`{"listing_id":%d}`, id)))
	modReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(modReq)
	if err != nil {
		// Модерация недоступна ИЛИ не ответила за отведённое время -
		// но объявление уже создано. Таймаут контекста (context.DeadlineExceeded)
		// и обычная сетевая ошибка размазаны тут в одну и ту же ветку:
		// это и есть статус "review", а не "pending", если причина - именно
		// таймаут. Различать их по ошибке в этом коде никто не пытается.
		if errors.Is(err, context.DeadlineExceeded) {
			listing.Status = "review"
		} else {
			listing.Status = "pending"
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": id, "status": listing.Status})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		listing.Status = "approved"
	} else {
		listing.Status = "rejected"
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id, "status": listing.Status})
}

// getListingHandler - GET /listings/{id}.
func getListingHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	listing, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listing)
}

// updatePriceHandler - PATCH /listings/{id}/price. Меняет цену, но
// только если объявление ещё не отклонено - бизнес-правило №4, зашитое
// прямо тут же, рядом с разбором запроса и ответом клиенту. Дальше -
// правило №5: смена цены может откатить approved обратно в pending,
// если новая цена выглядит подозрительной для этого продавца.
func updatePriceHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req updatePriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Price <= 0 {
		http.Error(w, "price must be positive", http.StatusBadRequest)
		return
	}

	listing, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Бизнес-правило №4: отклонённое объявление нельзя редактировать.
	if listing.Status == "rejected" {
		http.Error(w, "cannot update rejected listing", http.StatusConflict)
		return
	}

	listing.Price = req.Price

	// Бизнес-правило №5: если объявление уже approved, а новая цена
	// подозрительна для этого продавца - нужна повторная модерация,
	// статус откатывается в pending. Для review/pending ничего особого
	// не происходит - цена меняется, статус остаётся как был.
	if listing.Status == "approved" {
		seller, ok := sellers[listing.SellerID]
		if !ok {
			seller = &storedSeller{ID: listing.SellerID, Tier: "regular"}
		}
		priceLimit := 10_000_000.0
		if seller.Tier == "trusted" {
			priceLimit = 50_000_000.0
		}
		if req.Price > priceLimit {
			listing.Status = "pending"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listing)
}

// findByID - вспомогательная функция, имитирует поход в БД через
// database/sql, чтобы в коде было видно характерное ветвление
// sql.ErrNoRows - даже хотя реальной БД тут нет. Не используется
// напрямую в хендлерах выше (они читают db map напрямую), но
// пригодится как пример того, что нужно вынести в репозиторий.
func findByID(ctx context.Context, conn *sql.DB, id int64) (*storedListing, error) {
	row := conn.QueryRowContext(ctx, `SELECT * FROM listings WHERE id = $1`, id)
	var l storedListing
	err := row.Scan(&l.ID, &l.Title, &l.Price, &l.PhotoURL, &l.SellerID, &l.Status, &l.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Ошибка "не найдено" отдаётся вызывающему коду как есть -
		// кто вызвал findByID, тот и решает, что с sql.ErrNoRows делать.
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}
