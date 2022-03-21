package admin

import (
	graphqlhandler "github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/admin/transport"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource, auth config.AdminAPIAuth) *httproute.Router {
	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, p.RootHandler(newHealthzHandler))

	securityMiddleware := httproute.Chain(
		web.StaticSecurityHeadersMiddleware{},
		web.StaticCSPMiddleware{
			CSPDirectives: []string{
				"script-src 'self' 'unsafe-inline' cdn.jsdelivr.net",
				"object-src 'none'",
				"base-uri 'none'",
				"block-all-mixed-content",
				"frame-ancestors 'none'",
			},
		},
	)

	chain := httproute.Chain(
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoCache),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(func(p *deps.RequestProvider) httproute.Middleware {
			return newAuthorizationMiddleware(p, auth)
		}),
		httputil.CheckContentType([]string{
			graphqlhandler.ContentTypeJSON,
			graphqlhandler.ContentTypeGraphQL,
		}),
	)

	route := httproute.Route{Middleware: chain}

	router.AddRoutes(p.Handler(newGraphQLHandler), transport.ConfigureGraphQLRoute(route)...)
	router.Add(transport.ConfigurePresignImagesUploadRoute(route), p.Handler(newPresignImagesUploadHandler))

	return router
}
