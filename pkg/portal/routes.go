package portal

import (
	"net/http"

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

	rootChain := httproute.Chain(
		p.Middleware(newRecoverMiddleware),
		p.Middleware(newSentryMiddleware),
		p.Middleware(newSessionInfoMiddleware),
	)

	rootRoute := httproute.Route{Middleware: rootChain}

	router.Add(transport.ConfigureRuntimeConfigRoute(rootRoute), p.Handler(newRuntimeConfigHandler))
	router.Add(transport.ConfigureGraphQLRoute(rootRoute), p.Handler(newGraphQLHandler))
	router.Add(transport.ConfigureAdminAPIRoute(rootRoute), p.Handler(newAdminAPIHandler))

	return router
}
