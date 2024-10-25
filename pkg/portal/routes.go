package portal

import (
	"net/http"

	graphqlhandler "github.com/graphql-go/handler"

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
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.XFrameOptionsDeny),
		httproute.MiddlewareFunc(httputil.XRobotsTag),
		httputil.StaticCSPHeader{
			CSPDirectives: httputil.CSPDirectives{
				// FIXME(regeneratorRuntime)
				// parcel-2.0.0-rc.0 requires us to use ES6 module when the browser supports it.
				// ES6 module assumes strict mode.
				// regeneratorRuntime is not compatible with strict mode because
				// it uses Function to generate function, which is considered as eval.
				httputil.CSPDirective{
					Name: httputil.CSPDirectiveNameScriptSrc,
					Value: httputil.CSPSources{
						httputil.CSPSourceSelf,
						httputil.CSPSourceUnsafeEval,
						httputil.CSPSourceUnsafeInline,
						httputil.CSPHostSource{
							Host: "cdn.jsdelivr.net",
						},
						httputil.CSPHostSource{
							Host: "unpkg.com",
						},
						httputil.CSPHostSource{
							Host: "www.googletagmanager.com",
						},
						httputil.CSPHostSource{
							Host: "cdn.mxpnl.com",
						},
						httputil.CSPHostSource{
							Host: "eu.posthog.com",
						},
						httputil.CSPHostSource{
							Host: "eu-assets.i.posthog.com",
						},
						httputil.CSPHostSource{
							Host: "cmp.osano.com",
						},
					},
				},
				// monaco editor create worker with blob:
				httputil.CSPDirective{
					Name: httputil.CSPDirectiveNameWorkerSrc,
					Value: httputil.CSPSources{
						httputil.CSPSourceSelf,
						httputil.CSPSourceUnsafeInline,
						httputil.CSPHostSource{
							Host: "cdn.jsdelivr.net",
						},
						httputil.CSPSchemeSource{
							Scheme: "blob",
						},
					},
				},
				httputil.CSPDirective{
					Name: httputil.CSPDirectiveNameObjectSrc,
					Value: httputil.CSPSources{
						httputil.CSPSourceNone,
					},
				},
				httputil.CSPDirective{
					Name: httputil.CSPDirectiveNameBaseURI,
					Value: httputil.CSPSources{
						httputil.CSPSourceNone,
					},
				},
				httputil.CSPDirective{
					Name: httputil.CSPDirectiveNameBlockAllMixedContent,
				},
				// This must be kept in sync with httputil.XFrameOptionsDeny
				httputil.CSPDirective{
					Name: httputil.CSPDirectiveNameFrameAncestors,
					Value: httputil.CSPSources{
						httputil.CSPSourceNone,
					},
				},
			},
		},
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
	)

	rootChain := httproute.Chain(
		p.Middleware(newPanicMiddleware),
		p.Middleware(newBodyLimitMiddleware),
		p.Middleware(newSentryMiddleware),
		p.Middleware(newSessionInfoMiddleware),
		securityMiddleware,
	)
	systemConfigJSONChain := httproute.Chain(
		rootChain,
		httproute.MiddlewareFunc(httputil.NoCache),
	)
	graphqlChain := httproute.Chain(
		rootChain,
		httproute.MiddlewareFunc(httputil.NoStore),
		httputil.CheckContentType([]string{
			graphqlhandler.ContentTypeJSON,
			graphqlhandler.ContentTypeGraphQL,
		}),
	)
	adminAPIChain := httproute.Chain(
		rootChain,
		httproute.MiddlewareFunc(httputil.NoStore),
		p.Middleware(newSessionRequiredMiddleware),
		httputil.CheckContentType([]string{
			graphqlhandler.ContentTypeJSON,
			graphqlhandler.ContentTypeGraphQL,
		}),
	)
	incomingWebhookChain := httproute.Chain(
		rootChain,
		httproute.MiddlewareFunc(httputil.NoStore),
	)
	notFoundChain := httproute.Chain(
		securityMiddleware,
		httputil.GzipMiddleware{},
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

	return router
}
