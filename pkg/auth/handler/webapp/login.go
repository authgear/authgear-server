package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigureLoginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/login")
}

type LoginProvider interface {
	GetLoginForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	LoginWithLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
	LoginIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (func(error), error)
}

type LoginHandler struct {
	Provider  LoginProvider
	DBContext db.Context
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.DBContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetLoginForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("x_idp_id") != "" {
				writeResponse, err := h.Provider.LoginIdentityProvider(w, r, r.Form.Get("x_idp_id"))
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.LoginWithLoginID(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})

	return
}
