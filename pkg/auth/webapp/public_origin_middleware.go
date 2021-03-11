package webapp

import (
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type PublicOriginMiddleware struct {
	Config     *config.HTTPConfig
	TrustProxy config.TrustProxy
}

func (m *PublicOriginMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicOrigin, err := url.Parse(m.Config.PublicOrigin)
		if err != nil {
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

		http.Redirect(w, r, newURL.String(), http.StatusTemporaryRedirect)
	})
}
