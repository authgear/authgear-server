package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureRootRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/")
}

type RootHandler struct{}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u := *r.URL
	q := u.Query()
	webapp.RemoveX(q)
	u.RawQuery = q.Encode()
	u.Path = "/login"
	http.Redirect(w, r, httputil.HostRelative(&u).String(), http.StatusFound)
}
