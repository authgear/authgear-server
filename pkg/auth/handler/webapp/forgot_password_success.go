package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/db"
)

func ConfigureForgotPasswordSuccessHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/forgot_password/success").
		Methods("OPTIONS", "GET").
		Handler(h)
}

type ForgotPasswordSuccessProvider interface {
	GetForgotPasswordSuccess(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ForgotPasswordSuccessHandler struct {
	Provider  ForgotPasswordSuccessProvider
	DBContext db.Context
}

func (h *ForgotPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.DBContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetForgotPasswordSuccess(w, r)
			writeResponse(err)
			return err
		}
		return nil
	})
}
