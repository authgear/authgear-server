package admin

import (
	"net/http"

	graphqlhandler "github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/admin/transport"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httproute/httprouteotel"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource, auth config.AdminAPIAuth) http.Handler {
	router := httprouteotel.NewOTelRouter(httproute.NewRouter())

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, p.RootHandler(newHealthzHandler))

	chain := httproute.Chain(
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),

		httproute.MiddlewareFunc(httputil.NoStore),
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.XRobotsTag),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
		// x-frame-options: deny must be kept in sync of content-security-policy.
		httproute.MiddlewareFunc(httputil.XFrameOptionsDeny),
		// content-security-policy must be kept in sync of x-frame-options.
		httproute.MiddlewareFunc(AdminCSPMiddleware),

		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		// The following middlewares are project-specific.
		p.Middleware(func(p *deps.RequestProvider) httproute.Middleware {
			return newAuthorizationMiddleware(p, auth)
		}),
		p.Middleware(newUIParamMiddleware),

		// The following middlewares may terminate the request,
		// so they are ordered just before the handler, to make sure
		// the middlewares above always write their headers.
		httputil.CheckContentType([]string{
			graphqlhandler.ContentTypeJSON,
			graphqlhandler.ContentTypeGraphQL,
		}),
	)

	route := httproute.Route{Middleware: chain}

	router.AddRoutes(p.Handler(newGraphQLHandler), transport.ConfigureGraphQLRoute(route)...)
	router.Add(transport.ConfigurePresignImagesUploadRoute(route), p.Handler(newPresignImagesUploadHandler))
	router.Add(transport.ConfigureUserImportCreateRoute(route), p.Handler(newUserImportCreateHandler))
	router.Add(transport.ConfigureUserImportGetRoute(route), p.Handler(newUserImportGetHandler))
	router.Add(transport.ConfigureUserExportCreateRoute(route), p.Handler(newUserExportCreateHandler))
	router.Add(transport.ConfigureUserExportGetRoute(route), p.Handler(newUserExportGetHandler))

	return router.HTTPHandler()
}
