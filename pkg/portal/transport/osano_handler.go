package transport

import (
	"net/http"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureOsanoRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("GET").WithPathPattern("/api/osano.js")
}

type OsanoHandler struct {
	OsanoConfig *portalconfig.OsanoConfig
}

func (h *OsanoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.OsanoConfig.ScriptSrc != "" {
		http.Redirect(w, r, h.OsanoConfig.ScriptSrc, http.StatusFound)
	}
	// Otherwise we return an empty script.
	w.Header().Set("Content-Type", "application/javascript")
}
