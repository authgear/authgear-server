package sentry

import (
	"net/http"

	"github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/httputil"
)

type Middleware struct {
	SentryHub    *sentry.Hub
	ServerConfig *config.ServerConfig
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := m.SentryHub.Clone()
		hub.Scope().SetRequest(MakeMinimalSentryRequest(r, m.ServerConfig))
		r = r.WithContext(sentry.SetHubOnContext(r.Context(), hub))
		next.ServeHTTP(w, r)
	})
}

var HeaderWhiteList = []string{
	"Referer",
	"User-Agent",
	"X-Forwarded-For",
	"X-Real-IP",
	"Forwarded",
}

func MakeMinimalSentryRequest(r *http.Request, serverCfg *config.ServerConfig) (req *http.Request) {
	u := *r.URL
	u.Scheme = httputil.GetProto(r, serverCfg.TrustProxy)
	u.Host = httputil.GetHost(r, serverCfg.TrustProxy)

	req, _ = http.NewRequest(r.Method, u.String(), nil)

	for _, name := range HeaderWhiteList {
		if header := r.Header.Get(name); header != "" {
			req.Header.Set(name, header)
		}
	}

	return
}
