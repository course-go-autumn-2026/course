package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			startedAt := time.Now()
			response := newResponseWriter(writer)
			next.ServeHTTP(response, request)

			level := slog.LevelInfo
			if response.status >= http.StatusInternalServerError {
				level = slog.LevelError
			}
			logger.LogAttrs(
				request.Context(),
				level,
				"request handled",
				slog.String("method", request.Method),
				slog.String("route", routeTemplate(request)),
				slog.Int("status", response.status),
				slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
			)
		})
	}
}
