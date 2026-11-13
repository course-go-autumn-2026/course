package middleware

import (
	"context"
	"net/http"
	"time"
)

type metricRecorder interface {
	RequestStarted(context.Context, string, string)
	RequestFinished(context.Context, string, string, int, float64)
}

func Metrics(recorder metricRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			route := routeTemplate(request)
			recorder.RequestStarted(request.Context(), request.Method, route)
			startedAt := time.Now()
			response := newResponseWriter(writer)
			next.ServeHTTP(response, request)
			recorder.RequestFinished(
				request.Context(),
				request.Method,
				routeTemplate(request),
				response.status,
				time.Since(startedAt).Seconds(),
			)
		})
	}
}
