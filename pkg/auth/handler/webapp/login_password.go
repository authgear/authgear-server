package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
)

func AttachLoginPasswordHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/login/password").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newLoginPasswordHandler))
}

type loginPasswordProvider interface {
	GetLoginPasswordForm(w http.ResponseWriter, r *http.Request) (func(err error), error)
	PostLoginPassword(w http.ResponseWriter, r *http.Request) (func(err error), error)
}

type LoginPasswordHandler struct {
	Provider loginPasswordProvider
}

func (h *LoginPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		writeResponse, err := h.Provider.GetLoginPasswordForm(w, r)
		writeResponse(err)
		return
	}

	if r.Method == "POST" {
		writeResponse, err := h.Provider.PostLoginPassword(w, r)
		writeResponse(err)
		return
	}
}
