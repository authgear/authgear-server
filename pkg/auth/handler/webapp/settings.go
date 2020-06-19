package webapp

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/deps"
)

func AttachSettingsHandler(
	router *mux.Router,
	p *deps.RootProvider,
) {
	router.
		NewRoute().
		Path("/settings").
		Handler(p.Handler(newSettingsHandler))
}

type SettingsHandler struct {
	RenderProvider webapp.RenderProvider
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUISettingsHTML, nil)
}
