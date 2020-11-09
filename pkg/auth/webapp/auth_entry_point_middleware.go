package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type AuthEntryPointMiddleware struct {
	TrustProxy config.TrustProxy
}

func (m AuthEntryPointMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := session.GetUserID(r.Context())
		webSession := GetSession(r.Context())

		hasPrompt := false
		if webSession != nil {
			hasPrompt = webSession.Prompt != ""
		}

		if userID != nil && !hasPrompt {
			defaultRedirectURI := "/settings"
			redirectURI := GetRedirectURI(r, bool(m.TrustProxy), defaultRedirectURI)

			http.Redirect(w, r, redirectURI, http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
