package middleware

import (
	"net/http"

	getsentry "github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type SentryMiddleware struct {
	SentryHub    *getsentry.Hub
	ServerConfig *config.ServerConfig
}

func (m *SentryMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := m.SentryHub.Clone()
		hub.Scope().SetRequest(sentry.MakeMinimalRequest(r, m.ServerConfig.TrustProxy))
		r = r.WithContext(getsentry.SetHubOnContext(r.Context(), hub))
		next.ServeHTTP(w, r)
	})
}
