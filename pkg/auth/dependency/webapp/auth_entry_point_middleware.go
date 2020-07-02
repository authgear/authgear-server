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
		requireLoginPrompt := r.URL.Query().Get("prompt") == "login"
		if user != nil && !requireLoginPrompt {
			RedirectToRedirectURI(w, r, m.ServerConfig.TrustProxy)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
