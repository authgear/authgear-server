package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type PublicOriginMiddlewareLogger struct{ *log.Logger }

func NewPublicOriginMiddlewareLogger(lf *log.Factory) PublicOriginMiddlewareLogger {
	return PublicOriginMiddlewareLogger{lf.New("public-origin-middleware")}
}

type PublicOriginMiddleware struct {
	Config     *config.HTTPConfig
	TrustProxy config.TrustProxy
	Logger     PublicOriginMiddlewareLogger
}

func (m *PublicOriginMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicOrigin, err := url.Parse(m.Config.PublicOrigin)
		if err != nil {
			m.Logger.WithError(err).Error("failed to parse public origin")
			panic(err)
		}

		requestScheme := httputil.GetProto(r, bool(m.TrustProxy))
		requestHost := httputil.GetHost(r, bool(m.TrustProxy))

		if publicOrigin.Scheme == requestScheme && publicOrigin.Host == requestHost {
			next.ServeHTTP(w, r)
			return
		}

		newURL := *r.URL
		newURL.Scheme = publicOrigin.Scheme
		newURL.Host = publicOrigin.Host

		m.Logger.WithField("new_url", newURL).Info("redirect to the configured public origin")
		http.Redirect(w, r, newURL.String(), http.StatusTemporaryRedirect)
	})
}
