package siteadmin

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal"
	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/siteadmin/transport"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider) http.Handler {
	router := httproute.NewRouter()
	router.Health(p.Handler(newHealthzHandler))

	securityMiddleware := httproute.Chain(
		httproute.MiddlewareFunc(httputil.XContentTypeOptionsNosniff),
		httproute.MiddlewareFunc(httputil.XFrameOptionsDeny),
		httproute.MiddlewareFunc(httputil.XRobotsTag),
		httproute.MiddlewareFunc(portal.PortalCSPMiddleware),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
	)

	rootChain := httproute.Chain(
		p.Middleware(newOtelMiddleware),
		p.Middleware(newPanicMiddleware),
		p.Middleware(newBodyLimitMiddleware),
		p.Middleware(newSentryMiddleware),
	)

	apiChain := httproute.Chain(
		rootChain,
		// TODO: Authorization will be handled later
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoStore),
	)

	route := httproute.Route{Middleware: apiChain}
	router.AddRoutes(p.Handler(newProjectsListHandler), transport.ConfigureProjectsListRoute(route)...)

	return router.HTTPHandler()
}
