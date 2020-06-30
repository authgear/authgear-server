package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigureSignupRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/signup")
}

type SignupProvider interface {
	GetCreateLoginIDForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	CreateLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
	LoginIdentityProvider(w http.ResponseWriter, r *http.Request, providerAlias string) (func(error), error)
}

type SignupHandler struct {
	Provider  SignupProvider
	DBContext db.Context
}

func (h *SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.DBContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetCreateLoginIDForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			if r.Form.Get("x_idp_id") != "" {
				writeResponse, err := h.Provider.LoginIdentityProvider(w, r, r.Form.Get("x_idp_id"))
				writeResponse(err)
				return err
			}

			writeResponse, err := h.Provider.CreateLoginID(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
