package webapp

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
	"github.com/skygeario/skygear-server/pkg/httproute"
)

func ConfigureSettingsRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/settings")
}

type SettingsHandler struct {
	RenderProvider webapp.RenderProvider
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RenderProvider.WritePage(w, r, webapp.TemplateItemTypeAuthUISettingsHTML, nil)
}
