package webapp

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var PublicOriginMiddlewareLogger = slogutil.NewLogger("public-origin-middleware")

type PublicOriginMiddleware struct {
	Config     *config.HTTPConfig
	TrustProxy config.TrustProxy
}

func (m *PublicOriginMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := PublicOriginMiddlewareLogger.GetLogger(ctx)

		publicOrigin, err := url.Parse(m.Config.PublicOrigin)
		if err != nil {
			err = fmt.Errorf("failed to parse public origin: %w", err)
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

		logger.Debug(ctx, "redirect to the configured public origin", slog.String("new_url", newURL.String()))
		http.Redirect(w, r, newURL.String(), http.StatusTemporaryRedirect)
	})
}
