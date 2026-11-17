package middleware

import "net/http"

type responseWriter struct {
	http.ResponseWriter
	status int
}

func newResponseWriter(writer http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: writer, status: http.StatusOK}
}

func (writer *responseWriter) WriteHeader(status int) {
	writer.status = status
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *responseWriter) Write(body []byte) (int, error) {
	return writer.ResponseWriter.Write(body)
}
