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

func NewRouter(p *deps.RootProvider) *httproute.Router {
	router := httproute.NewRouter()
	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	securityMiddleware := httproute.Chain(
		web.StaticSecurityHeadersMiddleware{},
		web.StaticCSPMiddleware{
			CSPDirectives: []string{
				// FIXME(regeneratorRuntime)
				// parcel-2.0.0-rc.0 requires us to use ES6 module when the browser supports it.
				// ES6 module assumes strict mode.
				// regeneratorRuntime is not compatible with strict mode because
				// it uses Function to generate function, which is considered as eval.
				"script-src 'self' 'unsafe-eval' 'unsafe-inline' cdn.jsdelivr.net",
				"worker-src 'self' 'unsafe-inline' cdn.jsdelivr.net",
				"object-src 'none'",
				"base-uri 'none'",
				"block-all-mixed-content",
				"frame-ancestors 'none'",
			},
		},
	)

	rootChain := httproute.Chain(
		p.Middleware(newPanicMiddleware),
		p.Middleware(newBodyLimitMiddleware),
		p.Middleware(newSentryMiddleware),
		p.Middleware(newSessionInfoMiddleware),
		securityMiddleware,
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

	router.Add(transport.ConfigureStripeWebhookRoute(rootRoute), p.Handler(newStripeWebhookHandler))

	router.NotFound(securityMiddleware.Handle(p.Handler(newStaticAssetsHandler)))

	return router
}
