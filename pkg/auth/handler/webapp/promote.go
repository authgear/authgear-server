package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigurePromoteRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/promote_user")
}

type PromoteProvider interface {
	GetPromoteLoginIDForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	PromoteLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
	PromoteIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (func(error), error)
}

type PromoteHandler struct {
	Provider  PromoteProvider
	DBContext db.Context
}

func (h *PromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.DBContext.WithTx(func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetPromoteLoginIDForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("x_idp_id") != "" {
				writeResponse, err := h.Provider.PromoteIdentityProvider(w, r, r.Form.Get("x_idp_id"))
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.PromoteLoginID(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
