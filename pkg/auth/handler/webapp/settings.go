package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func AttachSettingsHandler(
	router *mux.Router,
	authDependency auth.DependencyMap,
) {
	router.
		NewRoute().
		Path("/settings").
		Handler(auth.MakeHandler(authDependency, newSettingsHandler))
}

type SettingsHandler struct {
	RenderProvider webapp.RenderProvider
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUISettingsHTML, nil)
}
