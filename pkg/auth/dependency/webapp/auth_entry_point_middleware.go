package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
)

type AuthEntryPointMiddleware struct {
	ServerConfig *config.ServerConfig
}

func (m AuthEntryPointMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUser(r.Context())
		hasState := r.URL.Query().Get("x_sid") != ""
		if user != nil && !hasState {
			redirectURI := GetRedirectURI(r, m.ServerConfig.TrustProxy)
			http.Redirect(w, r, redirectURI, http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
