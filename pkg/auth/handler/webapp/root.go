package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureRootRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/")
}

type RootHandler struct {
	AuthenticationConfig *config.AuthenticationConfig
	Cookies              CookieManager
	SignedUpCookie       webapp.SignedUpCookieDef
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	signedUpCookie, err := h.Cookies.GetCookie(r, h.SignedUpCookie.Def)
	signedUp := (err == nil && signedUpCookie.Value == "true")

	path := "/signup"
	if h.AuthenticationConfig.PublicSignupDisabled || signedUp {
		path = "/login"
	}

	u := webapp.MakeRelativeURL(path, webapp.PreserveQuery(r.URL.Query()))
	httputil.Redirect(r.Context(), w, r, u.String(), http.StatusFound)
}
