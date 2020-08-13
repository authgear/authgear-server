package deps

import (
	"net/http"

	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var utilsDeps = wire.NewSet(
	wire.NewSet(
		NewCookieFactory,
		wire.Bind(new(session.CookieFactory), new(*httputil.CookieFactory)),
		wire.Bind(new(newinteraction.CookieFactory), new(*httputil.CookieFactory)),
		wire.Bind(new(webapp.CookieFactory), new(*httputil.CookieFactory)),
	),
)

func NewCookieFactory(r *http.Request, serverConfig *config.ServerConfig) *httputil.CookieFactory {
	return &httputil.CookieFactory{
		Request:    r,
		TrustProxy: serverConfig.TrustProxy,
	}
}
