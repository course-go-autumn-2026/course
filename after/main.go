package main

import (
	"log"
	"net/http"

	"github.com/nestera-sh/course_06/handler"
	"github.com/nestera-sh/course_06/repository"
	"github.com/nestera-sh/course_06/service"
)

func main() {
	// Инициализация адаптеров
	repo := repository.NewMemoryListingRepo()
	moderator := &repository.HTTPModerator{URL: "http://moderation.internal/check"}
	antifraud := &repository.HTTPAntifraud{URL: "http://antifraud.internal/check"}

	// Внедрение зависимостей (DI) в сервис
	svc := service.NewListingService(repo, moderator, antifraud)

	// Инициализация хендлера
	h := handler.NewListingHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /listings", h.Publish)
	mux.HandleFunc("GET /listings/{id}", h.Get)
	mux.HandleFunc("PATCH /listings/{id}/price", h.UpdatePrice)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
