package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AuthEntryPointMiddleware struct {
	ServerConfig *config.ServerConfig
}

func (m AuthEntryPointMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserID(r.Context())
		hasState := r.URL.Query().Get("x_sid") != ""
		if userID != nil && !hasState {
			redirectURI := GetRedirectURI(r, m.ServerConfig.TrustProxy)
			http.Redirect(w, r, redirectURI, http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
