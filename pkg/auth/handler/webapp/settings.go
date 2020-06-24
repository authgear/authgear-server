package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

func ConfigureSettingsHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/settings").
		Methods("OPTIONS", "GET").
		Handler(h)
}

type SettingsHandler struct {
	RenderProvider webapp.RenderProvider
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUISettingsHTML, nil)
}
