package resolver

import (
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/resolver/handler"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource) *httproute.Router {
	router := httproute.NewRouter()

	rootChain := httproute.Chain(
		p.RootMiddleware(newPanicEndMiddleware),
		p.RootMiddleware(newPanicWriteEmptyResponseMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newPanicLogMiddleware),
	)

	chain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionMiddleware),
	)

	rootRoute := httproute.Route{Middleware: rootChain}
	route := httproute.Route{Middleware: chain}

	router.Add(rootRoute.WithMethods("GET").WithPathPattern("/healthz"), p.Handler(newHealthzHandler))
	router.AddRoutes(p.Handler(newSessionResolveHandler), handler.ConfigureResolveRoute(route)...)

	return router
}
