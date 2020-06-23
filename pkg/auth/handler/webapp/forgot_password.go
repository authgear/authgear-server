package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/db"
)

func ConfigureForgotPasswordHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/forgot_password").
		Methods("OPTIONS", "POST", "GET").
		Handler(h)
}

type forgotPasswordProvider interface {
	GetForgotPasswordForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	PostForgotPasswordForm(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ForgotPasswordHandler struct {
	Provider  forgotPasswordProvider
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
