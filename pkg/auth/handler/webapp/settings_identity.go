package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func AttachSettingsIdentityHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/settings/identity").
		Handler(auth.MakeHandler(authDependency, newSettingsIdentityHandler))
}

type SettingsIdentityHandler struct {
	RenderProvider webapp.RenderProvider
}

func (h *SettingsIdentityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUISettingsIdentityHTML, nil)
}
