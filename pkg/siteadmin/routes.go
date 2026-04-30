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
		p.Middleware(newCORSMiddleware),
	)

	apiChain := httproute.Chain(
		rootChain,
		p.Middleware(newSessionInfoMiddleware),
		p.Middleware(newAuthzMiddleware),
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoStore),
	)

	route := httproute.Route{Middleware: apiChain}
	router.Add(transport.ConfigureAppsListRoute(route), p.Handler(newAppsListHandler))
	router.Add(transport.ConfigureAppGetRoute(route), p.Handler(newAppGetHandler))
	router.Add(transport.ConfigureCollaboratorsListRoute(route), p.Handler(newCollaboratorsListHandler))
	router.Add(transport.ConfigureCollaboratorAddRoute(route), p.Handler(newCollaboratorAddHandler))
	router.Add(transport.ConfigureCollaboratorRemoveRoute(route), p.Handler(newCollaboratorRemoveHandler))
	router.Add(transport.ConfigureCollaboratorPromoteRoute(route), p.Handler(newCollaboratorPromoteHandler))
	router.Add(transport.ConfigureMessagingUsageRoute(route), p.Handler(newMessagingUsageHandler))
	router.Add(transport.ConfigureMonthlyActiveUsersUsageRoute(route), p.Handler(newMonthlyActiveUsersUsageHandler))
	router.Add(transport.ConfigurePlansListRoute(route), p.Handler(newPlansListHandler))
	router.Add(transport.ConfigureAppPlanChangeRoute(route), p.Handler(newAppPlanChangeHandler))

	return router.HTTPHandler()
}
