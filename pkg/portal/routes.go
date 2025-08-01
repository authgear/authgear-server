package portal

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/transport"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider) http.Handler {
	router := httproute.NewRouter()
	router.Health(p.Handler(newHealthzHandler))

	securityMiddleware := httproute.Chain(
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.XFrameOptionsDeny),
		httproute.MiddlewareFunc(httputil.XRobotsTag),
		httproute.MiddlewareFunc(PortalCSPMiddleware),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
	)

	rootChain := httproute.Chain(
		p.Middleware(newOtelMiddleware),
		p.Middleware(newPanicMiddleware),
		p.Middleware(newBodyLimitMiddleware),
		p.Middleware(newSentryMiddleware),
	)
	systemConfigJSONChain := httproute.Chain(
		rootChain,
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoCache),
	)
	graphqlChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionInfoMiddleware),
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoStore),
		httproute.MiddlewareFunc(httputil.CheckContentType([]string{
			graphqlutil.ContentTypeJSON,
			graphqlutil.ContentTypeGraphQL,
		})),
	)
	adminAPIChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionInfoMiddleware),
		// Middlewares that write headers are intentionally left out for this chain.
		// It is because the handler of this chain is a httputil.ReverseProxy.
		// We assume the proxied response has correct headers.
	)
	incomingWebhookChain := httproute.Chain(
		rootChain,
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoStore),
	)
	notFoundChain := httproute.Chain(
		securityMiddleware,
	)

	systemConfigJSONRoute := httproute.Route{Middleware: systemConfigJSONChain}
	graphqlRoute := httproute.Route{Middleware: graphqlChain}
	adminAPIRoute := httproute.Route{Middleware: adminAPIChain}
	incomingWebhookRoute := httproute.Route{Middleware: incomingWebhookChain}
	notFoundRoute := httproute.Route{Middleware: notFoundChain}

	router.Add(transport.ConfigureSystemConfigRoute(systemConfigJSONRoute), p.Handler(newSystemConfigHandler))

	router.Add(transport.ConfigureGraphQLRoute(graphqlRoute), p.Handler(newGraphQLHandler))

	router.Add(transport.ConfigureAdminAPIRoute(adminAPIRoute), p.Handler(newAdminAPIHandler))

	router.Add(transport.ConfigureStripeWebhookRoute(incomingWebhookRoute), p.Handler(newStripeWebhookHandler))

	router.Add(transport.ConfigureOsanoRoute(notFoundRoute), p.Handler(newOsanoHandler))

	router.NotFound(notFoundRoute, p.Handler(newStaticAssetsHandler))

	return router.HTTPHandler()
}
