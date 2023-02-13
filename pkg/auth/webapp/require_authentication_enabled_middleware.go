package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type RequireAuthenticationEnabledMiddleware struct {
	AuthUI *config.UIConfig
}

func (m RequireAuthenticationEnabledMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticationDisabled := m.AuthUI.AuthenticationDisabled
		if authenticationDisabled {
			http.Redirect(w, r, "/errors/feature_disabled", http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
