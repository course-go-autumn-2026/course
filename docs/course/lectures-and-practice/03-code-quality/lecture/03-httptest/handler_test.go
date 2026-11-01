package httptestdemo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGreetingHandler(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		target       string
		wantStatus   int
		wantGreeting string
	}{
		{
			name:         "success",
			method:       http.MethodGet,
			target:       "/greet?name=Alice",
			wantStatus:   http.StatusOK,
			wantGreeting: "Hello, Alice!",
		},
		{
			name:       "missing name",
			method:     http.MethodGet,
			target:     "/greet",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			target:     "/greet?name=Alice",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			recorder := httptest.NewRecorder()

			GreetingHandler(recorder, req)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.wantGreeting == "" {
				return
			}

			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			var response greetingResponse
			require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
			assert.Equal(t, tt.wantGreeting, response.Greeting)
		})
	}
}
