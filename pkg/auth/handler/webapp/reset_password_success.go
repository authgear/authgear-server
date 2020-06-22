package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/db"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func AttachResetPasswordSuccessHandler(
	router *mux.Router,
	p *deps.RootProvider,
) {
	router.
		NewRoute().
		Path("/reset_password/success").
		Methods("OPTIONS", "GET").
		Handler(p.Handler(newResetPasswordSuccessHandler))
}

type resetPasswordSuccessProvider interface {
	GetResetPasswordSuccess(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ResetPasswordSuccessHandler struct {
	Provider  resetPasswordSuccessProvider
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
