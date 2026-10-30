// Package before - версия "до": весь путь публикации объявления
// в одной функции. Так начинает выглядеть почти любой сервис,
// если не следить за границами.
package before

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type PublishListingRequest struct {
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	PhotoURL string  `json:"photo_url"`
	SellerID int64   `json:"seller_id"`
}

// PublishListingHandler делает всё сразу: разбирает запрос,
// проверяет бизнес-правила, пишет в БД, зовёт модерацию
// и решает, что ответить клиенту.
func PublishListingHandler(db *sql.DB, moderationURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PublishListingRequest
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

		// Поход в базу прямо из хендлера.
		var id int64
		err := db.QueryRowContext(r.Context(),
			`INSERT INTO listings (title, price, photo_url, seller_id, status, created_at)
			 VALUES ($1, $2, $3, $4, 'pending', $5) RETURNING id`,
			req.Title, req.Price, req.PhotoURL, req.SellerID, time.Now(),
		).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Поход во внешний сервис модерации прямо из хендлера.
		modReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
			moderationURL+"/check", strings.NewReader(`{"listing_id":`+itoa(id)+`}`))
		modReq.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(modReq)
		if err != nil {
			// Модерация недоступна - но объявление уже создано в БД.
			// Что с ним теперь делать? Решаем тут же, по месту.
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{"id": id, "status": "pending"})
			return
		}
		defer resp.Body.Close()

		status := "pending"
		if resp.StatusCode == http.StatusOK {
			status = "approved"
			_, _ = db.ExecContext(r.Context(),
				`UPDATE listings SET status = $1 WHERE id = $2`, status, id)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": id, "status": status})
	}
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
