package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigureEnterLoginIDRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/enter_login_id")
}

type EnterLoginIDProvider interface {
	GetEnterLoginIDForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	EnterLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
	RemoveLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type EnterLoginIDHandler struct {
	Provider  EnterLoginIDProvider
	DBContext db.Context
}

func (h *EnterLoginIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.DBContext.WithTx(func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetEnterLoginIDForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("x_action") == "remove" {
				writeResponse, err := h.Provider.RemoveLoginID(w, r)
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.EnterLoginID(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
