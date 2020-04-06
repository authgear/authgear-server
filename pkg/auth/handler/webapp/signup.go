package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
)

func AttachSignupHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/signup").
		Methods("OPTIONS", "POST", "GET").
		Handler(auth.MakeHandler(authDependency, newSignupHandler))
}

type signupProvider interface {
	GetSignupForm(w http.ResponseWriter, r *http.Request) (func(error), error)
	PostSignupLoginID(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type SignupHandler struct {
	Provider signupProvider
}

func (h *SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		writeResponse, err := h.Provider.GetSignupForm(w, r)
		writeResponse(err)
		return
	}

	if r.Method == "POST" {
		writeResponse, err := h.Provider.PostSignupLoginID(w, r)
		writeResponse(err)
		return
	}
}
