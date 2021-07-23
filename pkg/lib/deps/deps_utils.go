package deps

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/nonce"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	oidchandler "github.com/authgear/authgear-server/pkg/lib/oauth/oidc/handler"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var utilsDeps = wire.NewSet(
	wire.NewSet(
		httputil.DependencySet,
		NewCookieManager,
		wire.Bind(new(session.CookieManager), new(*httputil.CookieManager)),
		wire.Bind(new(idpsession.CookieManager), new(*httputil.CookieManager)),
		wire.Bind(new(idpsession.ResolverCookieManager), new(*httputil.CookieManager)),
		wire.Bind(new(oauth.ResolverCookieManager), new(*httputil.CookieManager)),
		wire.Bind(new(oidchandler.CookieManager), new(*httputil.CookieManager)),
		wire.Bind(new(interaction.CookieManager), new(*httputil.CookieManager)),
		wire.Bind(new(httputil.FlashMessageCookieManager), new(*httputil.CookieManager)),
		wire.Bind(new(nonce.CookieManager), new(*httputil.CookieManager)),
	),
)

func NewCookieManager(
	r *http.Request,
	trustProxy config.TrustProxy,
	httpCfg *config.HTTPConfig,
) *httputil.CookieManager {
	m := &httputil.CookieManager{
		CookiePrefix: httpCfg.CookiePrefix,
		Request:      r,
		TrustProxy:   bool(trustProxy),
	}
	if httpCfg.CookieDomain != nil {
		m.CookieDomain = *httpCfg.CookieDomain
	}
	return m
}
