package portal

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/transport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider, staticAsset StaticAssetConfig) *httproute.Router {
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

	sessionRequiredChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionRequiredMiddleware),
	)

	rootRoute := httproute.Route{Middleware: rootChain}
	sessionRequiredRoute := httproute.Route{Middleware: sessionRequiredChain}

	router.Add(transport.ConfigureRuntimeConfigRoute(rootRoute), p.Handler(newRuntimeConfigHandler))
	// It is OK to access portal graphql endpoint without session.
	// Actually the client check if viewer is null to determine session existence.
	router.Add(transport.ConfigureGraphQLRoute(rootRoute), p.Handler(newGraphQLHandler))

	// FIXME(portal): Access control to admin API.
	router.Add(transport.ConfigureAdminAPIRoute(sessionRequiredRoute), p.Handler(newAdminAPIHandler))

	if staticAsset.ServingEnabled {
		router.NotFound(http.FileServer(&httputil.TryFileSystem{
			Fallback: "/index.html",
			FS:       http.Dir(staticAsset.Directory),
		}))
	}

	return router
}
