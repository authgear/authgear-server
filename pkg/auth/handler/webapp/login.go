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
		Handler(auth.MakeHandler(authDependency, newLoginHandler))
}

type loginProvider interface {
	Login(w http.ResponseWriter, r *http.Request) (func(error), error)
}

type LoginHandler struct {
	LoginProvider loginProvider
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeResponse, err := h.LoginProvider.Login(w, r)
	writeResponse(err)
}
