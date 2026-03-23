package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureRobotsTXTRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods(http.MethodGet, http.MethodHead).
		WithPathPattern("/robots.txt")
}

type RobotsTXTHandler struct {
	Renderer Renderer
}

func (h *RobotsTXTHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	h.Renderer.Render(w, r, TemplateWebRobotsTXT, nil)
}
