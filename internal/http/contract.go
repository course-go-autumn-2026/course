//go:generate go tool mockgen -source contract.go -destination mocks_test.go -package httptransport_test
//go:generate go tool mockgen -destination handler_mocks_test.go -package httptransport_test net/http Handler

package httptransport

import (
	"net/http"

	"github.com/DaleCoopTP/grishin_lab1/internal/generated"
)

type createHandler interface {
	Handle(w http.ResponseWriter, r *http.Request, params api.CreateTripParams)
}

type tripHandler interface {
	Handle(w http.ResponseWriter, r *http.Request, id api.TripId)
}
