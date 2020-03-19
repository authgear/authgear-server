package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func AttachRootHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/").
		Handler(auth.MakeHandler(authDependency, newRootHandler))
}

type RootHandler struct {
	AuthenticateProvider webapp.AuthenticateProvider
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.AuthenticateProvider.ServeHTTP(w, r)
}
