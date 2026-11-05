package finish_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/DaleCoopTP/grishin_lab1/internal/domain/trip"
	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/trip/finish"
)

func TestHandler(t *testing.T) {
	id := uuid.New()
	cases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"success", nil, 200, ""},
		{"missing", trip.ErrNotFound, 404, "trip_not_found"},
		{"completed", trip.ErrCompleted, 409, "trip_completed"},
		{"internal", errors.New("private database details"), 500, "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			usecase := NewMockusecase(ctrl)
			now := time.Date(2026, time.November, 1, 12, 0, 0, 0, time.UTC)
			result := trip.Trip{ID: id, UserID: uuid.New(), DriverID: uuid.New(), Price: 1450, Status: trip.StatusCompleted, StartedAt: now}
			if result.Status == trip.StatusCompleted {
				result.FinishedAt = &now
			}
			usecase.EXPECT().Finish(gomock.Any(), id).DoAndReturn(func(ctx context.Context, got uuid.UUID) (trip.Trip, error) {
				if ctx.Err() != nil {
					t.Fatal(ctx.Err())
				}
				return result, tc.err
			})
			writer := response.New(slog.New(slog.NewTextHandler(io.Discard, nil)))
			handler := finish.New(usecase, writer)
			req := httptest.NewRequest("POST", "/api/v1/trips/"+id.String(), nil)
			out := httptest.NewRecorder()
			handler.Handle(out, req, id)
			if out.Code != tc.status {
				t.Fatalf("status %d, body %s", out.Code, out.Body.String())
			}
			if tc.err == nil {
				var body api.Trip
				if err := json.Unmarshal(out.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Id != id || body.Price != 1450 || string(body.Status) != result.Status ||
					(body.FinishedAt == nil) != (result.FinishedAt == nil) || out.Header().Get("Content-Type") != "application/json" {
					t.Fatalf("unexpected trip response: %+v", body)
				}
			} else {
				var body api.Problem
				if err := json.Unmarshal(out.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Code != tc.code || body.Instance == nil || *body.Instance != req.URL.Path ||
					out.Header().Get("Content-Type") != "application/problem+json" || strings.Contains(out.Body.String(), "private") {
					t.Fatalf("unexpected Problem: %s", out.Body.String())
				}
			}
		})
	}
}
