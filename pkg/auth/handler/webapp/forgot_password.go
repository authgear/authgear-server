package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
)

func AttachForgotPasswordHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/forgot_password").
		Handler(auth.MakeHandler(authDependency, newForgotPasswordHandler))
}

type forgotPasswordProvider interface {
	GetForgotPasswordForm(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type ForgotPasswordHandler struct {
	Provider forgotPasswordProvider
}

func (h *ForgotPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		writeResponse, err := h.Provider.GetForgotPasswordForm(w, r)
		writeResponse(err)
		return
	}
}
