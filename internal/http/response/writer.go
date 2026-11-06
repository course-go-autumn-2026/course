package response

import (
	"encoding/json"
	"net/http"

	"github.com/DaleCoopTP/grishin_lab1/internal/http/domainerror"
)

type Writer struct{}

func New() *Writer {
	return &Writer{}
}

func (w *Writer) Error(out http.ResponseWriter, r *http.Request, err error) {
	status, body := domainerror.Problem(err)
	path := r.URL.Path
	body.Instance = &path
	data, marshalErr := json.Marshal(body)
	if marshalErr != nil {
		http.Error(out, "Internal server error", http.StatusInternalServerError)
		return
	}
	out.Header().Set("Content-Type", "application/problem+json")
	out.WriteHeader(status)
	_, _ = out.Write(data)
}

func (w *Writer) JSON(out http.ResponseWriter, r *http.Request, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		w.Error(out, r, err)
		return
	}
	out.Header().Set("Content-Type", "application/json")
	out.WriteHeader(status)
	_, _ = out.Write(data)
}
