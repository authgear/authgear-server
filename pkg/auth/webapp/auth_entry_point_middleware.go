package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type AuthEntryPointMiddleware struct {
	ServerConfig *config.ServerConfig
}

func (m AuthEntryPointMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := session.GetUserID(r.Context())
		hasState := r.URL.Query().Get("x_sid") != ""
		if userID != nil && !hasState {
			redirectURI := GetRedirectURI(r, m.ServerConfig.TrustProxy)
			http.Redirect(w, r, redirectURI, http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
