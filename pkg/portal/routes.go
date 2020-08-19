package portal

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/upstreamapp"
	"github.com/authgear/authgear-server/pkg/portal/deps"
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

	router.Add(rootRoute.WithMethods("GET").WithPathPattern("/api"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionInfo := upstreamapp.GetValidSessionInfo(r.Context())
		if sessionInfo == nil {
			w.Write([]byte("No session"))
		} else {
			w.Write([]byte(fmt.Sprintf("User ID: %v", sessionInfo.UserID)))
		}
	}))

	return router
}
