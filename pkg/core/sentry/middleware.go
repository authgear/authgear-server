package sentry

import (
	"net"
	"net/http"
	"net/url"

	"github.com/getsentry/sentry-go"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

func Middleware(hub *sentry.Hub) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hub := hub.Clone()
			hub.Scope().SetRequest(MakeMinimalSentryRequest(r))
			r = r.WithContext(sentry.SetHubOnContext(r.Context(), hub))
			next.ServeHTTP(w, r)
		})
	}
}

var HeaderWhiteList = []string{
	"Referer",
	"User-Agent",
	"X-Forwarded-For",
	"X-Real-IP",
	"Forwarded",
}

func MakeMinimalSentryRequest(r *http.Request) (req sentry.Request) {
	req.Method = r.Method

	url := url.URL{}
	url.Scheme = corehttp.GetProto(r)
	url.Host = corehttp.GetHost(r)
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
