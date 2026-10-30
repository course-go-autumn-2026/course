package main

import (
	"log"
	"net/http"

	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/handler"
	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/repository"
	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/repository/client"
	"github.com/course-go-autumn-2026/practice6-variant2-booking/internal/service"
)

const (
	notifyURL           = "http://notify.internal/send"
	roomAvailabilityURL = "http://room-availability.internal/check"
)

func main() {
	// 1. Инициализируем репозитории и внешние клиенты (адаптеры)
	bookingRepo := repository.NewMemoryBookingRepository()
	employeeRepo := repository.NewMemoryEmployeeRepository()
	availabilityClient := client.NewRoomAvailabilityClient(roomAvailabilityURL)
	notifierClient := client.NewNotificationClient(notifyURL)

	// 2. Инициализируем сервисный слой
	bookingService := service.NewBookingService(
		bookingRepo,
		employeeRepo,
		availabilityClient,
		notifierClient,
	)

	// 3. Инициализируем HTTP-хендлер и роутер
	bookingHandler := handler.NewBookingHandler(bookingService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, bookingHandler)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
