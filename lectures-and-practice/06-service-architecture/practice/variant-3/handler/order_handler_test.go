package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/course-go-autumn-2026/practice6-variant3-order/domain"
	"github.com/course-go-autumn-2026/practice6-variant3-order/repository"
	"github.com/course-go-autumn-2026/practice6-variant3-order/service"
)

type checkFunc func(context.Context) (bool, error)

func (f checkFunc) Check(ctx context.Context) (bool, error) { return f(ctx) }

func testMux(orders service.OrderRepository, blocked bool) http.Handler {
	stock := checkFunc(func(context.Context) (bool, error) { return true, nil })
	fraud := checkFunc(func(context.Context) (bool, error) { return blocked, nil })
	svc := service.NewOrderService(orders, repository.NewMemoryCustomerRepository(), stock, fraud)
	h := NewOrderHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", h.Create)
	mux.HandleFunc("GET /orders/{id}", h.Get)
	mux.HandleFunc("PATCH /orders/{id}/items", h.UpdateItems)
	return mux
}

func TestHTTPValidation(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body string
		code                     int
		text                     string
	}{
		{"empty", "POST", "/orders", `{"items":[]}`, 400, "order must have at least one item\n"},
		{"quantity", "POST", "/orders", `{"items":[{"quantity":0,"price":500}]}`, 400, "item quantity must be positive\n"},
		{"price", "POST", "/orders", `{"items":[{"quantity":1,"price":0}]}`, 400, "item price must be positive\n"},
		{"regular_limit", "POST", "/orders", `{"customer_id":1,"items":[{"quantity":1,"price":100001}]}`, 400, "order total exceeds limit for this customer\n"},
		{"new_limit", "POST", "/orders", `{"customer_id":999,"items":[{"quantity":1,"price":100001}]}`, 400, "order total exceeds limit for this customer\n"},
		{"vip_limit", "POST", "/orders", `{"customer_id":2,"items":[{"quantity":1,"price":1000001}]}`, 400, "order total exceeds limit for this customer\n"},
		{"invalid_json", "POST", "/orders", `{`, 400, ""},
		{"get_bad_id", "GET", "/orders/abc", "", 400, ""},
		{"get_missing", "GET", "/orders/999", "", 404, ""},
		{"patch_bad_id", "PATCH", "/orders/abc/items", `{}`, 400, ""},
		{"patch_bad_json", "PATCH", "/orders/1/items", `{`, 400, ""},
		{"empty_before_missing", "PATCH", "/orders/999/items", `{"items":[]}`, 400, "order must have at least one item\n"},
		{"patch_missing", "PATCH", "/orders/999/items", `{"items":[{"quantity":1,"price":1}]}`, 404, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mux := testMux(repository.NewMemoryOrderRepository(), false)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if w.Code != tc.code || w.Body.String() != tc.text {
				t.Fatalf("got %d %q, want %d %q", w.Code, w.Body.String(), tc.code, tc.text)
			}
		})
	}
}

func TestHTTPOrderLifecycle(t *testing.T) {
	mux := testMux(repository.NewMemoryOrderRepository(), false)
	created := httptest.NewRecorder()
	mux.ServeHTTP(created, httptest.NewRequest("POST", "/orders", strings.NewReader(
		`{"customer_id":2,"items":[{"sku":"sofa","quantity":2,"price":150000}]}`,
	)))
	var response map[string]any
	if err := json.Unmarshal(created.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if created.Code != 201 || len(response) != 2 || response["id"] != float64(1) || response["status"] != "confirmed" {
		t.Fatalf("unexpected create response: %d %s", created.Code, created.Body.String())
	}
	for _, tc := range []struct {
		method, path, body string
		total              float64
		status             domain.Status
	}{
		{"GET", "/orders/1", "", 300000, domain.StatusConfirmed},
		{"PATCH", "/orders/1/items", `{"items":[{"sku":"sofa","quantity":3,"price":500000}]}`, 1500000, domain.StatusPendingStock},
		{"GET", "/orders/1", "", 1500000, domain.StatusPendingStock},
	} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
		var order domain.Order
		if err := json.Unmarshal(w.Body.Bytes(), &order); err != nil {
			t.Fatal(err)
		}
		if w.Code != 200 || w.Header().Get("Content-Type") != "application/json" || order.Total != tc.total || order.Status != tc.status || order.ConfirmedQty != 2 {
			t.Fatalf("unexpected order response: %d %s", w.Code, w.Body.String())
		}
		var fields map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
			t.Fatal(err)
		}
		if len(fields) != 7 || fields["CustomerID"] != float64(2) || fields["Items"] == nil || fields["CreatedAt"] == nil {
			t.Fatalf("order JSON fields changed: %v", fields)
		}
	}
}

func TestFraudBlockedHTTP(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	mux := testMux(repo, true)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/orders", strings.NewReader(
		`{"customer_id":1,"items":[{"sku":"lamp","quantity":1,"price":500}]}`,
	)))
	if w.Code != 403 || w.Body.String() != "payment looks fraudulent\n" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
	id, err := repo.Create(context.Background(), &domain.Order{})
	if err != nil || id != 1 {
		t.Fatalf("blocked request consumed an ID: %d, %v", id, err)
	}
}

func TestCancelledOrderHTTP(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	id, err := repo.Create(context.Background(), &domain.Order{Status: domain.StatusCancelled})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	testMux(repo, false).ServeHTTP(w, httptest.NewRequest("PATCH", fmt.Sprintf("/orders/%d/items", id), strings.NewReader(
		`{"items":[{"sku":"lamp","quantity":1,"price":500}]}`,
	)))
	if w.Code != 409 || w.Body.String() != "cannot update cancelled order\n" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}

func TestInternalErrorHTTP(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, errors.New("private database details"))
	if w.Code != 500 || w.Body.String() != "internal error\n" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}
