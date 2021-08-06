package portal

import (
	"net/http"

	graphqlhandler "github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/transport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider, staticAsset StaticAssetConfig) *httproute.Router {
	router := httproute.NewRouter()
	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	// TODO(csp): improve security
	secMiddleware := &web.SecHeadersMiddleware{
		CSPDirectives: []string{
			"script-src 'self' 'unsafe-inline' cdn.jsdelivr.net",
			"worker-src 'self' 'unsafe-inline' cdn.jsdelivr.net",
			"object-src 'none'",
			"base-uri 'none'",
			"block-all-mixed-content",
		},
	}

	rootChain := httproute.Chain(
		p.Middleware(newPanicEndMiddleware),
		p.Middleware(newPanicWriteEmptyResponseMiddleware),
		p.Middleware(newPanicLogMiddleware),
		p.Middleware(newBodyLimitMiddleware),
		p.Middleware(newSentryMiddleware),
		p.Middleware(newSessionInfoMiddleware),
		secMiddleware,
		httproute.MiddlewareFunc(httputil.NoCache),
	)

	graphqlChain := httproute.Chain(
		rootChain,
		httputil.CheckContentType([]string{
			graphqlhandler.ContentTypeJSON,
			graphqlhandler.ContentTypeGraphQL,
		}),
	)

	adminAPIChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionRequiredMiddleware),
		httputil.CheckContentType([]string{
			graphqlhandler.ContentTypeJSON,
			graphqlhandler.ContentTypeGraphQL,
		}),
	)

	rootRoute := httproute.Route{Middleware: rootChain}
	graphqlRoute := httproute.Route{Middleware: graphqlChain}
	adminAPIRoute := httproute.Route{Middleware: adminAPIChain}

	router.Add(transport.ConfigureSystemConfigRoute(rootRoute), p.Handler(newSystemConfigHandler))
	router.Add(transport.ConfigureGraphQLRoute(graphqlRoute), p.Handler(newGraphQLHandler))

	router.Add(transport.ConfigureAdminAPIRoute(adminAPIRoute), p.Handler(newAdminAPIHandler))

	if staticAsset.ServingEnabled {
		router.NotFound(secMiddleware.Handle(p.Handler(newStaticAssetsHandler)))
	}

	return router
}
