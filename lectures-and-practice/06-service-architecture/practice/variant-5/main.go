// Package main - сервис модерации комментариев, три ручки, весь код
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

type storedComment struct {
	ID        int64
	AuthorID  int64
	Text      string
	Status    string // "pending" | "published" | "hidden"
	CreatedAt time.Time
}

// storedAuthor - автор. Карта вместо настоящей таблицы authors, но
// влияет на бизнес-правила комментария точно так же, как влияла бы
// запись из реальной БД.
type storedAuthor struct {
	ID         int64
	Reputation string // "new" | "trusted"
}

var (
	db     = map[int64]*storedComment{}
	nextID int64

	// Авторы захардкожены прямо тут - для практики достаточно двух:
	// нового и доверенного. author_id, которого нет в карте, считается
	// новым ("new") по умолчанию.
	authors = map[int64]*storedAuthor{
		1: {ID: 1, Reputation: "new"},
		2: {ID: 2, Reputation: "trusted"},
	}

	spamFilterURL    = "http://spam-filter.internal/check"
	toxicityCheckURL = "http://toxicity-check.internal/score"
)

// --- Запросы/ответы ---

type createCommentRequest struct {
	AuthorID int64  `json:"author_id"`
	Text     string `json:"text"`
}

type updateTextRequest struct {
	Text string `json:"text"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /comments", createCommentHandler)
	mux.HandleFunc("GET /comments/{id}", getCommentHandler)
	mux.HandleFunc("PATCH /comments/{id}/text", updateTextHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// createCommentHandler - POST /comments. Делает всё сразу: валидация,
// спам-фильтр, репутация автора, порог токсичности по ней, поход в
// сервис проверки токсичности, запись в "базу", решение по ответу.
func createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Бизнес-правило №1: пустой комментарий смысла не имеет.
	if strings.TrimSpace(req.Text) == "" {
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}

	// Спам-фильтр - ДО проверки токсичности и ДО записи в базу. В
	// отличие от проверки токсичности, недоступность спам-фильтра не
	// блокирует публикацию и не переводит комментарий в особый статус
	// - считаем проверку пройденной и просто идём дальше (fail-open,
	// а не fail-safe, как у токсичности).
	spReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
		spamFilterURL, strings.NewReader(fmt.Sprintf(`{"text":%q}`, req.Text)))
	spReq.Header.Set("Content-Type", "application/json")
	isSpam := false
	spResp, spErr := http.DefaultClient.Do(spReq)
	if spErr == nil {
		defer spResp.Body.Close()
		if spResp.StatusCode == http.StatusForbidden {
			isSpam = true
		}
	}
	// spErr != nil (спам-фильтр недоступен) - isSpam остаётся false, едем дальше.

	if isSpam {
		http.Error(w, "comment looks like spam", http.StatusBadRequest)
		return
	}

	// Автора смотрим прямо тут же, через ту же map, что и комментарии.
	// Неизвестный author_id - не ошибка, просто "new" по умолчанию.
	author, ok := authors[req.AuthorID]
	if !ok {
		author = &storedAuthor{ID: req.AuthorID, Reputation: "new"}
	}

	nextID++
	id := nextID
	comment := &storedComment{
		ID:        id,
		AuthorID:  req.AuthorID,
		Text:      req.Text,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	db[id] = comment

	// Поход в сервис проверки токсичности прямо из хендлера. Сервис
	// отвечает либо 200 (не токсично), либо 403 (токсично) - но что
	// делать с результатом "не токсично", зависит от репутации автора:
	// у доверенного это сразу published, у нового - всё равно pending,
	// публикация нового автора без отдельного шага модерации не
	// предусмотрена вообще, независимо от токсичности.
	txReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
		toxicityCheckURL, strings.NewReader(fmt.Sprintf(`{"text":%q}`, comment.Text)))
	txReq.Header.Set("Content-Type", "application/json")
	txResp, txErr := http.DefaultClient.Do(txReq)
	if txErr != nil {
		// Сервис проверки токсичности недоступен - fail-safe:
		// комментарий остаётся pending, а не публикуется "на всякий случай".
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": id, "status": comment.Status})
		return
	}
	defer txResp.Body.Close()

	if txResp.StatusCode == http.StatusOK {
		if author.Reputation == "trusted" {
			comment.Status = "published"
		}
		// у нового автора остаётся pending даже при "не токсично".
	} else {
		comment.Status = "hidden"
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id, "status": comment.Status})
}

// getCommentHandler - GET /comments/{id}.
func getCommentHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	comment, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// updateTextHandler - PATCH /comments/{id}/text. Меняет текст, но
// только если комментарий не скрыт - бизнес-правило №4, зашитое прямо
// тут же, рядом с разбором запроса и ответом клиенту. Дальше -
// правило №5: если комментарий был published, новый текст заново
// проходит проверку токсичности, и при превышении порога статус
// откатывается в pending.
func updateTextHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req updateTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Text) == "" {
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}

	comment, ok := db[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Бизнес-правило №4: скрытый комментарий нельзя редактировать.
	if comment.Status == "hidden" {
		http.Error(w, "cannot edit a hidden comment", http.StatusConflict)
		return
	}

	comment.Text = req.Text

	// Бизнес-правило №5: если комментарий был published, новый текст
	// заново проверяется на токсичность - при превышении порога статус
	// откатывается в pending. Для pending ничего особого не происходит
	// - текст просто меняется, статус остаётся как был.
	if comment.Status == "published" {
		txReq, _ := http.NewRequestWithContext(r.Context(), http.MethodPost,
			toxicityCheckURL, strings.NewReader(fmt.Sprintf(`{"text":%q}`, req.Text)))
		txReq.Header.Set("Content-Type", "application/json")
		txResp, txErr := http.DefaultClient.Do(txReq)
		if txErr != nil {
			// Сервис недоступен при редактировании - на всякий случай
			// уводим на повторную модерацию, а не оставляем published.
			comment.Status = "pending"
		} else {
			defer txResp.Body.Close()
			if txResp.StatusCode != http.StatusOK {
				comment.Status = "pending"
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// findByID - вспомогательная функция, имитирует поход в БД через
// database/sql, чтобы в коде было видно характерное ветвление
// sql.ErrNoRows - даже хотя реальной БД тут нет. Не используется
// напрямую в хендлерах выше (они читают db map напрямую), но
// пригодится как пример того, что нужно вынести в репозиторий.
func findByID(ctx context.Context, conn *sql.DB, id int64) (*storedComment, error) {
	row := conn.QueryRowContext(ctx, `SELECT * FROM comments WHERE id = $1`, id)
	var c storedComment
	err := row.Scan(&c.ID, &c.AuthorID, &c.Text, &c.Status, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Ошибка "не найдено" отдаётся вызывающему коду как есть -
		// кто вызвал findByID, тот и решает, что с sql.ErrNoRows делать.
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
