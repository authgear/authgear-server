package images

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/images/deps"
	"github.com/authgear/authgear-server/pkg/images/handler"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httproute/httprouteotel"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource) http.Handler {
	router := httprouteotel.NewOTelRouter(httproute.NewRouter())
	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, http.HandlerFunc(httputil.HealthCheckHandler))

	rootChain := httproute.Chain(
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
		httproute.MiddlewareFunc(httputil.XRobotsTag),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(newCORSMiddleware),
	)

	rootRoute := httproute.Route{Middleware: rootChain}
	router.Add(handler.ConfigureGetRoute(rootRoute), p.Handler(newGetHandler))
	router.Add(handler.ConfigurePostRoute(rootRoute), p.Handler(newPostHandler))

	return router.HTTPHandler()
}
