package sentry

import (
	"net"
	"net/http"
	"net/url"

	"github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/httputil"
)

type Middleware struct {
	Hub          *sentry.Hub
	ServerConfig *config.ServerConfig
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub := m.Hub.Clone()
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

func MakeMinimalSentryRequest(r *http.Request, serverCfg *config.ServerConfig) (req sentry.Request) {
	req.Method = r.Method

	url := url.URL{}
	url.Scheme = httputil.GetProto(r, serverCfg.TrustProxy)
	url.Host = httputil.GetHost(r, serverCfg.TrustProxy)
	url.Path = r.URL.Path
	req.URL = url.String()

	req.Headers = map[string]string{}
	for _, name := range HeaderWhiteList {
		if header := r.Header.Get(name); header != "" {
			req.Headers[name] = header
		}
	}

	if addr, port, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		req.Env = map[string]string{"REMOTE_ADDR": addr, "REMOTE_PORT": port}
	}

	return
}
