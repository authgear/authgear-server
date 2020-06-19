package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func AttachRootHandler(
	router *mux.Router,
	p *deps.RootProvider,
) {
	router.
		NewRoute().
		Path("/").
		Handler(&RootHandler{})
}

type RootHandler struct{}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	webapp.RedirectToPathWithoutX(w, r, "/login")
}
