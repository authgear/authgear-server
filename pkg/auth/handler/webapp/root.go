package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureRootRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/")
}

type RootHandler struct {
	AuthenticationConfig *config.AuthenticationConfig
	SignedUpCookie       webapp.SignedUpCookieDef
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	signedUp, err := r.Cookie(h.SignedUpCookie.Def.Name)
	path := "/signup"
	if h.AuthenticationConfig.PublicSignupDisabled || (err == nil && signedUp.Value == "true") {
		path = "/login"
	}
	http.Redirect(w, r, path, http.StatusFound)
}
