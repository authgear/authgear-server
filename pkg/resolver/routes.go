package resolver

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/resolver/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func newAllSessionMiddleware(p *deps.RequestProvider) httproute.Middleware {
	return newSessionMiddleware(p)
}

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource) http.Handler {
	router := httproute.NewRouter()

	router.Health(p.RootHandler(newHealthzHandler))

	chain := httproute.Chain(
		p.RootMiddleware(newOtelMiddleware),
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
		httproute.MiddlewareFunc(httputil.XRobotsTag),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newAllSessionMiddleware),
	)

	route := httproute.Route{Middleware: chain}

	router.AddRoutes(p.Handler(newSessionResolveHandler), handler.ConfigureResolveRoute(route)...)

	return router.HTTPHandler()
}
