package repository_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/course-go-autumn-2026/practice6-variant3-order/repository"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type responseBody struct{ closed bool }

func (b *responseBody) Read([]byte) (int, error) { return 0, io.EOF }
func (b *responseBody) Close() error             { b.closed = true; return nil }

type checker interface {
	Check(context.Context) (bool, error)
}

func TestHTTPChecks(t *testing.T) {
	for _, tc := range []struct {
		name  string
		fraud bool
		code  int
		want  bool
	}{
		{"stock_ok", false, 200, true},
		{"stock_refused", false, 403, false},
		{"stock_error", false, 500, false},
		{"fraud_ok", true, 200, false},
		{"fraud_blocked", true, 403, true},
		{"fraud_error", true, 500, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &responseBody{}
			client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != "POST" || r.URL.String() != "http://checks.test/check" || r.Body != nil {
					t.Fatalf("unexpected external request: %+v", r)
				}
				return &http.Response{StatusCode: tc.code, Body: body, Header: make(http.Header)}, nil
			})}
			var c checker = repository.NewHTTPStockChecker("http://checks.test/check", client)
			if tc.fraud {
				c = repository.NewHTTPPaymentFraudChecker("http://checks.test/check", client)
			}
			result, err := c.Check(context.Background())
			if err != nil || result != tc.want || !body.closed {
				t.Fatalf("got result=%v, err=%v, closed=%v", result, err, body.closed)
			}
		})
	}
}

func TestHTTPChecks_TransportError(t *testing.T) {
	want := errors.New("network unavailable")
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return nil, want
	})}
	for _, c := range []checker{
		repository.NewHTTPStockChecker("http://checks.test/check", client),
		repository.NewHTTPPaymentFraudChecker("http://checks.test/check", client),
	} {
		if _, err := c.Check(context.Background()); !errors.Is(err, want) {
			t.Fatalf("want transport error, got %v", err)
		}
	}
}
