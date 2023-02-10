package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type RequireSettingsEnabledMiddleware struct {
	AuthUI *config.UIConfig
}

func (m RequireSettingsEnabledMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		settingsDisabled := m.AuthUI.SettingsDisabled
		if settingsDisabled {
			http.Redirect(w, r, "/errors/feature_disabled", http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
