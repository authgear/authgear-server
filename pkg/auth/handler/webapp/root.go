package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/httproute"
)

func ConfigureRootRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/")
}

type RootHandler struct{}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	redirectURI := webapp.MakeURLWithPathWithoutX(r.URL, "/login")
	http.Redirect(w, r, redirectURI, http.StatusFound)
}
