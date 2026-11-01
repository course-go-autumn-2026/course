package handler

import (
	"fmt"
	"net/http"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=handler.go -destination=mocks/greeter_mock.go -package=mocks

// TODO: объявите здесь минимальный интерфейс для зависимости обработчика.

type Handler struct {
	service *GreetingService
}

func NewHandler(service *GreetingService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	name := request.URL.Query().Get("name")
	if name == "" {
		http.Error(writer, "name is required", http.StatusBadRequest)
		return
	}

	greeting, err := h.service.Greet(request.Context(), name)
	if err != nil {
		http.Error(writer, "service unavailable", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(writer, greeting)
}
