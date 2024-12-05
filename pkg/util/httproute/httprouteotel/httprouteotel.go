package httprouteotel

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

type Router interface {
	Add(route httproute.Route, h http.Handler)
	AddRoutes(h http.Handler, routes ...httproute.Route)
	NotFound(route httproute.Route, h http.Handler)
	HTTPHandler() http.Handler
}

type OTelRouter struct {
	router Router
}

func NewOTelRouter(router Router) Router {
	// It is observed that otelhttp >=v0.56 <=v0.57 has memory leak that
	// causing memory usage increase over time.
	// See https://github.com/open-telemetry/opentelemetry-go-contrib/issues/6315
	// So let's stop using it.
	// return &OTelRouter{
	// 	router: router,
	// }
	return router
}

func (r *OTelRouter) wrapHandler(route httproute.Route, h http.Handler) http.Handler {
	// Use the path pattern as route tag.
	routeTag := route.PathPattern
	h = otelhttp.WithRouteTag(routeTag, h)
	return h
}

func (r *OTelRouter) Add(route httproute.Route, h http.Handler) {
	h = r.wrapHandler(route, h)
	r.router.Add(route, h)
}

func (r *OTelRouter) AddRoutes(h http.Handler, routes ...httproute.Route) {
	for _, route := range routes {
		r.Add(route, h)
	}
}

func (r *OTelRouter) NotFound(route httproute.Route, h http.Handler) {
	h = r.wrapHandler(route, h)
	r.router.NotFound(route, h)
}

func (r *OTelRouter) HTTPHandler() http.Handler {
	h := r.router.HTTPHandler()
	// The route serves at /
	operation := "/"
	h = otelhttp.NewHandler(h, operation,
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		otelhttp.WithPublicEndpoint(),
		// It is unnecessary to use otelhttp.WithServerName.
		// It uses HTTP Header Host by default.
	)
	return h
}
