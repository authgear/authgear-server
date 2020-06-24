package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func ConfigureRootHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/").
		Methods("OPTIONS", "GET").
		Handler(h)
}

type RootHandler struct{}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	webapp.RedirectToPathWithoutX(w, r, "/login")
}
