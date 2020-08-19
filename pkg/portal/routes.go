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
		// FIXME(portal): add sentry middleware.
		// We cannot add it now because it depends pkg/lib/config.ServerConfig.TrustProxy.
		p.Middleware(newSessionInfoMiddleware),
	)

	rootRoute := httproute.Route{Middleware: rootChain}

	router.Add(transport.ConfigureGraphQLRoute(rootRoute), p.Handler(newGraphQLHandler))
	return router
}
