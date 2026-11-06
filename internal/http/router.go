package httptransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/domainerror"
	"github.com/DaleCoopTP/grishin_lab1/internal/http/response"
)

type serverHandlers struct {
	create createHandler
	get    tripHandler
	finish tripHandler
	health http.Handler
	ready  http.Handler
}

var _ api.ServerInterface = (*serverHandlers)(nil)

func NewRouter(create createHandler, get, finish tripHandler,
	health, ready http.Handler, response *response.Writer,
) chi.Router {
	handlers := &serverHandlers{create: create, get: get, finish: finish, health: health, ready: ready}
	router := chi.NewRouter()
	api.HandlerWithOptions(handlers, api.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, _ error) {
			response.Error(w, r, domainerror.ErrInvalidRequest)
		},
	})
	return router
}

func (h *serverHandlers) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	h.create.Handle(w, r, params)
}

func (h *serverHandlers) GetTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	h.get.Handle(w, r, id)
}

func (h *serverHandlers) FinishTrip(w http.ResponseWriter, r *http.Request, id api.TripId) {
	h.finish.Handle(w, r, id)
}

func (h *serverHandlers) Health(w http.ResponseWriter, r *http.Request) {
	h.health.ServeHTTP(w, r)
}

func (h *serverHandlers) Ready(w http.ResponseWriter, r *http.Request) {
	h.ready.ServeHTTP(w, r)
}
