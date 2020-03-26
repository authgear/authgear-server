package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func AttachLogoutHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/logout").
		Handler(auth.MakeHandler(authDependency, newLogoutHandler))
}

type LogoutHandler struct {
	RenderProvider webapp.RenderProvider
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUILogoutHTML, nil)
}
