package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"

	api "github.com/Smoz1e/Bulatov_Ilya_Lab/internal/generated"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.LogAttrs(
						request.Context(),
						slog.LevelError,
						"recovered panic",
						slog.Any("panic", recovered),
						slog.String("stack", string(debug.Stack())),
					)
					writeInternalError(writer, request.URL.Path)
				}
			}()
			next.ServeHTTP(writer, request)
		})
	}
}

func writeInternalError(writer http.ResponseWriter, instance string) {
	detail := "Internal server error"
	writer.Header().Set("Content-Type", "application/problem+json")
	writer.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(writer).Encode(api.Problem{
		Code:     "internal_error",
		Detail:   &detail,
		Instance: &instance,
		Status:   http.StatusInternalServerError,
		Title:    "Internal error",
		Type:     "https://tripgo.example/problems/internal-error",
	}); err != nil {
		return
	}
}
