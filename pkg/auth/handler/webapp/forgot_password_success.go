package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func AttachForgotPasswordSuccessHandler(
	router *mux.Router,
	p *deps.RootProvider,
) {
	router.
		NewRoute().
		Path("/forgot_password/success").
		Methods("OPTIONS", "GET").
		Handler(p.Handler(newForgotPasswordSuccessHandler))
}

type forgotPasswordSuccessProvider interface {
	GetForgotPasswordSuccess(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ForgotPasswordSuccessHandler struct {
	Provider  forgotPasswordSuccessProvider
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
