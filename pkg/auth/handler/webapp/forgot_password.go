package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/db"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigureForgotPasswordRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/forgot_password")
}

type ForgotPasswordProvider interface {
	GetForgotPasswordForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	PostForgotPasswordForm(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ForgotPasswordHandler struct {
	Provider  ForgotPasswordProvider
	DBContext db.Context
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.DBContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetForgotPasswordForm(w, r)
			writeResponse(err)
			return err
		}

		if r.Method == "POST" {
			writeResponse, err := h.Provider.PostForgotPasswordForm(w, r)
			writeResponse(err)
			return err
		}

		return nil
	})
}
