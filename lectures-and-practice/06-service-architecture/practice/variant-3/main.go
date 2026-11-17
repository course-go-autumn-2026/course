package main

import (
	"log"
	"net/http"
	"os"

	"github.com/course-go-autumn-2026/practice6-variant3-order/handler"
	"github.com/course-go-autumn-2026/practice6-variant3-order/repository"
	"github.com/course-go-autumn-2026/practice6-variant3-order/service"
)

func main() {
	orders := repository.NewMemoryOrderRepository()
	customers := repository.NewMemoryCustomerRepository()
	stock := repository.NewHTTPStockChecker(envOrDefault("STOCK_CHECK_URL", "http://stock-check.internal/check"), http.DefaultClient)
	fraud := repository.NewHTTPPaymentFraudChecker(envOrDefault("PAYMENT_FRAUD_URL", "http://payment-fraud.internal/check"), http.DefaultClient)
	svc := service.NewOrderService(orders, customers, stock, fraud)
	h := handler.NewOrderHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", h.Create)
	mux.HandleFunc("GET /orders/{id}", h.Get)
	mux.HandleFunc("PATCH /orders/{id}/items", h.UpdateItems)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
