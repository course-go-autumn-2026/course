package httptransport

import (
	"maps"
	"net/http"
	"slices"
	"strings"

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
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, r, domainerror.ErrRouteNotFound)
	})
	methods := registeredMethods(router)
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.RawPath
		if path == "" {
			path = r.URL.Path
		}
		ctx := chi.NewRouteContext()
		allowed := make([]string, 0, len(methods))
		for _, method := range methods {
			ctx.Reset()
			if router.Match(ctx, method, path) {
				allowed = append(allowed, method)
			}
		}
		w.Header().Set("Allow", strings.Join(allowed, ", "))
		response.Error(w, r, domainerror.ErrMethodNotAllowed)
	})
	return router
}

func registeredMethods(router chi.Routes) []string {
	methods := make(map[string]struct{})
	_ = chi.Walk(router, func(method, _ string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		methods[method] = struct{}{}
		return nil
	})
	return slices.Sorted(maps.Keys(methods))
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
