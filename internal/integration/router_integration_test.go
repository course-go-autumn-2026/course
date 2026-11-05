//go:build integration

package integration_test

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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	httptransport "github.com/DaleCoopTP/grishin_lab1/internal/http"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/health"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/ready"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
	httpcreate "github.com/DaleCoopTP/grishin_lab1/internal/http/trip/create"
	httpfinish "github.com/DaleCoopTP/grishin_lab1/internal/http/trip/finish"
	httpget "github.com/DaleCoopTP/grishin_lab1/internal/http/trip/get"
	"github.com/DaleCoopTP/grishin_lab1/internal/infra/postgres/transaction"
	triprepository "github.com/DaleCoopTP/grishin_lab1/internal/repository/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/create"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/finish"
	"github.com/DaleCoopTP/grishin_lab1/internal/usecase/trip/get"
)

const jsonContentType = "application/json"

func TestTripHTTPContract(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is not set")
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
	router := testRouter(pool)

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

	for _, path := range []string{"/health", "/ready"} {
		response := request(http.MethodGet, path, "", "")
		if response.Code != http.StatusOK || response.Header().Get("Content-Type") != jsonContentType || strings.TrimSpace(response.Body.String()) != `{"status":"ok"}` {
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
		assertProblem(t, request(http.MethodPost, "/api/v1/trips", body, jsonContentType), http.StatusBadRequest, "invalid_request")
	}
	assertProblem(t, request(http.MethodPost, "/api/v1/trips", validBody, "text/plain"), http.StatusBadRequest, "invalid_request")
	assertProblem(t, request(http.MethodGet, "/api/v1/trips/not-a-uuid", "", ""), http.StatusBadRequest, "invalid_request")

	created := request(http.MethodPost, "/api/v1/trips", validBody, jsonContentType)
	if created.Code != http.StatusCreated || created.Header().Get("Content-Type") != jsonContentType {
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
	assertProblem(t, request(http.MethodPost, "/api/v1/trips", validBody, jsonContentType), http.StatusConflict, "driver_busy")
	if response := request(http.MethodGet, path, "", ""); response.Code != http.StatusOK || response.Header().Get("Content-Type") != jsonContentType {
		t.Fatalf("get: status %d, body %s", response.Code, response.Body.String())
	}
	finished := request(http.MethodPost, path+"/finish", "", "")
	if finished.Code != http.StatusOK || finished.Header().Get("Content-Type") != jsonContentType {
		t.Fatalf("finish: status %d, body %s", finished.Code, finished.Body.String())
	}
	if err := json.Unmarshal(finished.Body.Bytes(), &trip); err != nil || trip.Status != api.Completed || trip.FinishedAt == nil {
		t.Fatalf("finished trip: %+v, error %v", trip, err)
	}
	assertProblem(t, request(http.MethodPost, path+"/finish", "", ""), http.StatusConflict, "trip_completed")
	missing := "/api/v1/trips/" + uuid.NewString()
	assertProblem(t, request(http.MethodGet, missing, "", ""), http.StatusNotFound, "trip_not_found")
	assertProblem(t, request(http.MethodPost, missing+"/finish", "", ""), http.StatusNotFound, "trip_not_found")

	keyedDriverID := uuid.New()
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := pool.Exec(cleanupCtx, "DELETE FROM trips WHERE driver_id = $1", keyedDriverID); err != nil {
			t.Errorf("delete keyed trip: %v", err)
		}
	}()
	keyedBody := strings.Replace(validBody, driverID.String(), keyedDriverID.String(), 1)
	key := uuid.NewString()
	requestWithKey := func(body, header string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/trips", strings.NewReader(body))
		req.Header.Set("Content-Type", jsonContentType)
		req.Header.Set("Idempotency-Key", header)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	first := requestWithKey(keyedBody, key)
	second := requestWithKey(keyedBody, key)
	if first.Code != http.StatusCreated || second.Code != http.StatusOK || first.Body.String() != second.Body.String() ||
		first.Header().Get("Location") != second.Header().Get("Location") {
		t.Fatalf("idempotency responses: first %d %s, second %d %s", first.Code, first.Body.String(), second.Code, second.Body.String())
	}
	assertProblem(t, requestWithKey(strings.Replace(keyedBody, `"price":1450`, `"price":1451`, 1), key), http.StatusConflict, "idempotency_conflict")
	assertProblem(t, requestWithKey(strings.Replace(keyedBody, `"price":1450`, `"price":-1`, 1), key), http.StatusBadRequest, "invalid_request")
	assertProblem(t, requestWithKey(keyedBody, "invalid"), http.StatusBadRequest, "invalid_request")
	testUnavailableDatabase(t, ctx, databaseURL, path)
}

func testRouter(pool *pgxpool.Pool) http.Handler {
	tx := transaction.NewTxManager(pool, 5*time.Second)
	repository := triprepository.NewRepository(tx, 3*time.Second)
	createTrip := create.New(repository, tx)
	getTrip := get.New(repository)
	finishTrip := finish.New(repository, tx)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	writer := response.New(logger)
	return httptransport.NewRouter(httpcreate.New(createTrip, writer), httpget.New(getTrip, writer),
		httpfinish.New(finishTrip, writer), health.New(writer), ready.New(pool, 3*time.Second, writer), writer)
}

func testUnavailableDatabase(t *testing.T, ctx context.Context, databaseURL, path string) {
	t.Helper()
	closedPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	closedPool.Close()
	probeRouter := testRouter(closedPool)

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
		if response.Code != tc.status || response.Header().Get("Content-Type") != jsonContentType || strings.TrimSpace(response.Body.String()) != tc.body {
			t.Fatalf("closed pool %s: status %d, body %s", tc.path, response.Code, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	probeRouter.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	assertProblem(t, response, http.StatusInternalServerError, "internal_error")
}

func assertProblem(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status || response.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("problem response: status %d, content-type %q, body %s", response.Code, response.Header().Get("Content-Type"), response.Body.String())
	}
	var problem api.Problem
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil || problem.Code != code || int(problem.Status) != status ||
		problem.Type == "" || problem.Title == "" || problem.Detail == nil || problem.Instance == nil {
		t.Fatalf("problem: got %+v, error %v", problem, err)
	}
}
