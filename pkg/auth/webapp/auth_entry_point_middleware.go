package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type AuthEntryPointMiddleware struct {
	TrustProxy  config.TrustProxy
	OAuthConfig *config.OAuthConfig
	UIConfig    *config.UIConfig
}

func (m AuthEntryPointMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := session.GetUserID(r.Context())
		webSession := GetSession(r.Context())

		fromAuthzEndpoint := false
		if webSession != nil {
			// stay in the auth entry point if login is triggered by authz endpoint
			fromAuthzEndpoint = webSession.FromAuthzEndpoint
		}

		if userID != nil && !fromAuthzEndpoint {
			defaultRedirectURI := DerivePostLoginRedirectURIFromRequest(r, m.OAuthConfig, m.UIConfig)
			redirectURI := GetRedirectURI(r, bool(m.TrustProxy), defaultRedirectURI)

			http.Redirect(w, r, redirectURI, http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
