package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureOAuthEntrypointRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/_internals/oauth_entrypoint")
}

type OAuthEntrypointHandler struct{}

func (h *OAuthEntrypointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := webapp.MakeRelativeURL("/flows/select_account", webapp.PreserveQuery(r.URL.Query()))
	http.Redirect(w, r, u.String(), http.StatusFound)
}
