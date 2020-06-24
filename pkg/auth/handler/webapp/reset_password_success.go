package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/db"
)

func ConfigureResetPasswordSuccessHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/reset_password/success").
		Methods("OPTIONS", "GET").
		Handler(h)
}

type ResetPasswordSuccessProvider interface {
	GetResetPasswordSuccess(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ResetPasswordSuccessHandler struct {
	Provider  ResetPasswordSuccessProvider
	DBContext db.Context
}

func (h *ResetPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.DBContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetResetPasswordSuccess(w, r)
			writeResponse(err)
			return err
		}
		return nil
	})
}
