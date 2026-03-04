package superadmin

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal/deps"
	"github.com/authgear/authgear-server/pkg/portal/superadmin/transport"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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
		httproute.MiddlewareFunc(SuperadminCSPMiddleware),
		httproute.MiddlewareFunc(httputil.PermissionsPolicyHeader),
	)

	rootChain := httproute.Chain(
		p.Middleware(newOtelMiddleware),
		p.Middleware(newPanicMiddleware),
		p.Middleware(newBodyLimitMiddleware),
		p.Middleware(newSentryMiddleware),
	)

	graphqlChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionInfoMiddleware),
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoStore),
		httproute.MiddlewareFunc(httputil.CheckContentType([]string{
			graphqlutil.ContentTypeJSON,
			graphqlutil.ContentTypeGraphQL,
		})),
	)

	route := httproute.Route{Middleware: graphqlChain}
	router.AddRoutes(p.Handler(newGraphQLHandler), transport.ConfigureGraphQLRoute(route)...)

	return router.HTTPHandler()
}
