package admin

import (
	graphqlhandler "github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/admin/transport"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
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
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.XFrameOptionsDeny),
		httproute.MiddlewareFunc(httputil.XRobotsTag),
		httputil.StaticCSPHeader{
			CSPDirectives: httputil.CSPDirectives{
				httputil.CSPDirective{
					Name: httputil.CSPDirectiveNameScriptSrc,
					Value: httputil.CSPSources{
						httputil.CSPSourceSelf,
						httputil.CSPSourceUnsafeInline,
						httputil.CSPHostSource{
							Host: "unpkg.com",
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

	chain := httproute.Chain(
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoStore),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(func(p *deps.RequestProvider) httproute.Middleware {
			return newAuthorizationMiddleware(p, auth)
		}),
		p.Middleware(newUIParamMiddleware),
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

	return router
}
