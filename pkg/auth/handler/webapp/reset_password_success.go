package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
)

func AttachResetPasswordSuccessHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/reset_password/success").
		Methods("OPTIONS", "GET").
		Handler(auth.MakeHandler(authDependency, newResetPasswordSuccessHandler))
}

type resetPasswordSuccessProvider interface {
	GetResetPasswordSuccess(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ResetPasswordSuccessHandler struct {
	Provider resetPasswordSuccessProvider
}

func (h *ResetPasswordSuccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		writeResponse, err := h.Provider.GetResetPasswordSuccess(w, r)
		writeResponse(err)
		return
	}
}
