package ready_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/DaleCoopTP/grishin_lab1/internal/http/ready"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
)

func TestReadiness(t *testing.T) {
	for _, available := range []bool{true, false} {
		t.Run(map[bool]string{true: "available", false: "unavailable"}[available], func(t *testing.T) {
			var pingErr error
			if !available {
				pingErr = errors.New("connection refused")
			}
			checker := NewMockchecker(gomock.NewController(t))
			checker.EXPECT().Ping(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > time.Second {
					t.Fatal("readiness check must have a bounded context")
				}
				return pingErr
			})
			writer := response.New()
			handler := ready.New(checker, time.Second, writer)
			out := httptest.NewRecorder()
			var recordedErr error
			ctx := response.WithErrorRecorder(t.Context(), func(err error) { recordedErr = err })
			handler.ServeHTTP(out, httptest.NewRequestWithContext(ctx, "GET", "/ready", nil))
			if recordedErr != pingErr {
				t.Fatalf("recorded error = %v, want %v", recordedErr, pingErr)
			}
			status, body := 200, `{"status":"ok"}`
			if !available {
				status, body = 503, `{"status":"unavailable"}`
			}
			if out.Code != status || strings.TrimSpace(out.Body.String()) != body ||
				out.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("status %d, body %s", out.Code, out.Body.String())
			}
		})
	}
}
