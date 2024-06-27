package resolver

import (
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/resolver/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func newAllSessionMiddleware(p *deps.RequestProvider) httproute.Middleware {
	return newSessionMiddleware(p)
}

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource) *httproute.Router {
	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, p.RootHandler(newHealthzHandler))

	chain := httproute.Chain(
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newAllSessionMiddleware),
	)

	route := httproute.Route{Middleware: chain}

	router.AddRoutes(p.Handler(newSessionResolveHandler), handler.ConfigureResolveRoute(route)...)

	return router
}
