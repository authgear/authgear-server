package middleware

import (
	"net/http"

	getsentry "github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

type SentryMiddleware struct {
	SentryHub  *getsentry.Hub
	TrustProxy config.TrustProxy
}

func (m *SentryMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := m.SentryHub.Clone()
		hub.Scope().SetRequest(sentry.MakeMinimalRequest(r, bool(m.TrustProxy)))
		r = r.WithContext(getsentry.SetHubOnContext(r.Context(), hub))
		next.ServeHTTP(w, r)
	})
}
