package admin

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/admin/transport"
	configsource "github.com/authgear/authgear-server/pkg/lib/config/source"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource) *httproute.Router {
	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	chain := httproute.Chain(
		p.RootMiddleware(newRootRecoverMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newRequestRecoverMiddleware),
		p.Middleware(newAuthorizationMiddleware),
	)

	route := httproute.Route{Middleware: chain}

	router.Add(transport.ConfigureGraphQLRoute(route), p.Handler(newGraphQLHandler))

	return router
}
