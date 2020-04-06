package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
)

func AttachLoginHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/login").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newLoginHandler))
}

type loginProvider interface {
	GetLoginForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	PostLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type LoginHandler struct {
	Provider loginProvider
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		writeResponse, err := h.Provider.GetLoginForm(w, r)
		writeResponse(err)
		return
	}

	if r.Method == "POST" {
		writeResponse, err := h.Provider.PostLoginID(w, r)
		writeResponse(err)
		return
	}
}
