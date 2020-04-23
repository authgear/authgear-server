package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

func AttachForgotPasswordSuccessHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/forgot_password/success").
		Methods("OPTIONS", "GET").
		Handler(auth.MakeHandler(authDependency, newForgotPasswordSuccessHandler))
}

type forgotPasswordSuccessProvider interface {
	GetForgotPasswordSuccess(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ForgotPasswordSuccessHandler struct {
	Provider  forgotPasswordSuccessProvider
	TxContext db.TxContext
}

func (h *ForgotPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.WithTx(h.TxContext, func() error {
		if r.Method == "GET" {
			writeResponse, err := h.Provider.GetForgotPasswordSuccess(w, r)
			writeResponse(err)
			return err
		}
		return nil
	})
}
