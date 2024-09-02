package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureOAuthEntrypointRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/_internals/oauth_entrypoint")
}

type OAuthEntrypointEndpointsProvider interface {
	SelectAccountEndpointURL() *url.URL
}

type OAuthEntrypointHandler struct {
	Endpoints OAuthEntrypointEndpointsProvider
}

func (h *OAuthEntrypointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := webapp.MakeRelativeURL(
		h.Endpoints.SelectAccountEndpointURL().Path,
		webapp.PreserveQuery(r.URL.Query()),
	)
	http.Redirect(w, r, u.String(), http.StatusFound)
}
