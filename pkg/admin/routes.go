package admin

import (
	"net/http"

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
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	chain := httproute.Chain(
		p.RootMiddleware(newPanicEndMiddleware),
		p.RootMiddleware(newPanicWriteEmptyResponseMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		&web.SecHeadersMiddleware{
			CSPDirectives: web.DefaultStrictCSPDirectives,
		},
		httproute.MiddlewareFunc(httputil.NoCache),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newPanicLogMiddleware),
		p.Middleware(func(p *deps.RequestProvider) httproute.Middleware {
			return newAuthorizationMiddleware(p, auth)
		}),
		httputil.CheckContentType([]string{
			graphqlhandler.ContentTypeJSON,
			graphqlhandler.ContentTypeGraphQL,
		}),
	)

	route := httproute.Route{Middleware: chain}

	router.Add(transport.ConfigureGraphQLRoute(route), p.Handler(newGraphQLHandler))

	return router
}
