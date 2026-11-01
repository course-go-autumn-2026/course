// Пакет httptestdemo демонстрирует тестирование HTTP-обработчика без запуска реального сервера.
package httptestdemo

import (
	"encoding/json"
	"net/http"
	"strings"
)

type greetingResponse struct {
	Greeting string `json:"greeting"`
}

// GreetingHandler возвращает JSON-приветствие для имени из query-параметра.
func GreetingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(greetingResponse{Greeting: "Hello, " + name + "!"}); err != nil {
		return
	}
}
