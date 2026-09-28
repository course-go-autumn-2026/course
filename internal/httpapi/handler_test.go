package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DaleCoopTP/grishin_lab1/internal/database"
	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/DaleCoopTP/grishin_lab1/internal/trips"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTripHTTPContract(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	tx := database.NewTxManager(pool, 5*time.Second)
	service := trips.NewService(trips.NewRepository(tx, 3*time.Second), tx)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewHandler(service, pool, 3*time.Second, logger)

	driverID := uuid.New()
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := pool.Exec(cleanupCtx, "DELETE FROM trips WHERE driver_id = $1", driverID); err != nil {
			t.Errorf("delete test trips: %v", err)
		}
	}()
	validBody := fmt.Sprintf(`{"user_id":%q,"driver_id":%q,"start_point":{"latitude":59.9398,"longitude":30.3146},"end_point":{"latitude":59.929,"longitude":30.3626},"price":1450}`, uuid.New().String(), driverID.String())

	request := func(method, path, body, contentType string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	assertProblem := func(response *httptest.ResponseRecorder, status int, code string) {
		t.Helper()
		if response.Code != status || response.Header().Get("Content-Type") != "application/problem+json" {
			t.Fatalf("problem response: status %d, content-type %q, body %s", response.Code, response.Header().Get("Content-Type"), response.Body.String())
		}
		var problem api.Problem
		if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil || problem.Code != code || problem.Status != int32(status) ||
			problem.Type == "" || problem.Title == "" || problem.Detail == nil || problem.Instance == nil {
			t.Fatalf("problem: got %+v, error %v", problem, err)
		}
	}

	for _, path := range []string{"/health", "/ready"} {
		response := request(http.MethodGet, path, "", "")
		if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" || strings.TrimSpace(response.Body.String()) != `{"status":"ok"}` {
			t.Fatalf("%s: status %d, body %s", path, response.Code, response.Body.String())
		}
	}

	for _, body := range []string{
		`{"user_id":"bad"}`,
		`{"user_id":"00000000-0000-0000-0000-000000000000","driver_id":"00000000-0000-0000-0000-000000000000","start_point":{"latitude":0,"longitude":0},"end_point":{"latitude":0,"longitude":0},"price":0}`,
		strings.Replace(validBody, `"price":1450`, `"price":-1`, 1),
		strings.Replace(validBody, `"latitude":59.9398`, `"latitude":91`, 1),
		strings.Replace(validBody, `"latitude":59.9398,`, ``, 1),
		strings.Replace(validBody, `"price":1450`, `"price":1450,"unexpected":true`, 1),
	} {
		assertProblem(request(http.MethodPost, "/api/v1/trips", body, "application/json"), http.StatusBadRequest, "invalid_request")
	}
	assertProblem(request(http.MethodPost, "/api/v1/trips", validBody, "text/plain"), http.StatusBadRequest, "invalid_request")
	assertProblem(request(http.MethodGet, "/api/v1/trips/not-a-uuid", "", ""), http.StatusBadRequest, "invalid_request")

	created := request(http.MethodPost, "/api/v1/trips", validBody, "application/json")
	if created.Code != http.StatusCreated || created.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("create: status %d, body %s", created.Code, created.Body.String())
	}
	var trip api.Trip
	if err := json.Unmarshal(created.Body.Bytes(), &trip); err != nil || trip.Id == uuid.Nil || trip.Status != api.Active {
		t.Fatalf("created trip: %+v, error %v", trip, err)
	}
	path := "/api/v1/trips/" + trip.Id.String()
	if created.Header().Get("Location") != path {
		t.Fatalf("Location: got %q, want %q", created.Header().Get("Location"), path)
	}
	assertProblem(request(http.MethodPost, "/api/v1/trips", validBody, "application/json"), http.StatusConflict, "driver_busy")
	if response := request(http.MethodGet, path, "", ""); response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("get: status %d, body %s", response.Code, response.Body.String())
	}
	finished := request(http.MethodPost, path+"/finish", "", "")
	if finished.Code != http.StatusOK || finished.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("finish: status %d, body %s", finished.Code, finished.Body.String())
	}
	if err := json.Unmarshal(finished.Body.Bytes(), &trip); err != nil || trip.Status != api.Completed || trip.FinishedAt == nil {
		t.Fatalf("finished trip: %+v, error %v", trip, err)
	}
	assertProblem(request(http.MethodPost, path+"/finish", "", ""), http.StatusConflict, "trip_completed")
	missing := "/api/v1/trips/" + uuid.NewString()
	assertProblem(request(http.MethodGet, missing, "", ""), http.StatusNotFound, "trip_not_found")
	assertProblem(request(http.MethodPost, missing+"/finish", "", ""), http.StatusNotFound, "trip_not_found")

	closedPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	closedPool.Close()
	closedTx := database.NewTxManager(closedPool, 5*time.Second)
	closedService := trips.NewService(trips.NewRepository(closedTx, 3*time.Second), closedTx)
	probeRouter := NewHandler(closedService, closedPool, 3*time.Second, logger)
	for _, tc := range []struct {
		path   string
		status int
		body   string
	}{
		{path: "/health", status: http.StatusOK, body: `{"status":"ok"}`},
		{path: "/ready", status: http.StatusServiceUnavailable, body: `{"status":"unavailable"}`},
	} {
		response := httptest.NewRecorder()
		probeRouter.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if response.Code != tc.status || response.Header().Get("Content-Type") != "application/json" || strings.TrimSpace(response.Body.String()) != tc.body {
			t.Fatalf("closed pool %s: status %d, body %s", tc.path, response.Code, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	probeRouter.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	assertProblem(response, http.StatusInternalServerError, "internal_error")
}
