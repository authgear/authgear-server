package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slice"
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
	userID := session.GetUserID(r.Context())
	webSession := webapp.GetSession(r.Context())

	loginPrompt := false
	fromAuthzEndpoint := false
	if webSession != nil {
		// stay in the auth entry point if prompt = login
		loginPrompt = slice.ContainsString(webSession.Prompt, "login")
		fromAuthzEndpoint = webSession.ClientID != ""
	}

	path := ""
	if fromAuthzEndpoint && userID != nil && !loginPrompt {
		path = "/select_account"
	} else {
		signedUpCookie, err := r.Cookie(h.SignedUpCookie.Def.Name)
		signedUp := (err == nil && signedUpCookie.Value == "true")
		path = GetAuthenticationEndpoint(signedUp, h.AuthenticationConfig.PublicSignupDisabled)
	}

	http.Redirect(w, r, path, http.StatusFound)
}

func GetAuthenticationEndpoint(signedUp bool, publicSignupDisabled bool) string {
	path := "/signup"
	if publicSignupDisabled || signedUp {
		path = "/login"
	}

	return path
}
