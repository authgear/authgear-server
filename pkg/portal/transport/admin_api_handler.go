package transport

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"slices"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureAdminAPIRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET", "POST").WithPathPattern("/api/apps/:appid/*path")
}

type AdminAPIAuthzService interface {
	ListAuthorizedApps(ctx context.Context, userID string) ([]string, error)
}

type AdminAPIService interface {
	Director(ctx context.Context, appID string, p string, userID string, usage service.Usage) (func(*http.Request), error)
}

var AdminAPILogger = slogutil.NewLogger("admin-api-proxy")

// graphiQLProxiedPaths are the Admin API paths that serve the GraphiQL IDE as
// an HTML document when the method is GET.
// See pkg/admin/transport/handler_graphql.go, which registers both.
//
// These are values of the *path route parameter, not request paths: the
// /api/apps/:appid prefix is consumed by the route pattern. So a browser
// opening
//
//	https://portal.example.com/api/apps/QXBwOmFjY291bnRz/graphql
//
// arrives here with p == "/graphql".
var graphiQLProxiedPaths = []string{
	"/graphql",
	"/_api/admin/graphql",
}

// isGraphiQLDocumentRequest reports whether p, the *path route parameter, is
// the one case that must be served without access control checking.
//
// GraphiQL is opened by a top-level browser navigation, and the portal session
// is carried in the Authorization header rather than a cookie, so such a
// navigation cannot authenticate itself. The IDE authenticates on its own once
// loaded, and executes every query with POST.
//
// The exemption is deliberately limited to the exact GraphiQL document paths.
// GET on those paths only ever returns the GraphiQL HTML page; the rest of the
// Admin API includes GET endpoints that expose data and mint presigned upload
// URLs, and those must not be reachable without access control.
func isGraphiQLDocumentRequest(method string, p string) bool {
	if method != "GET" {
		return false
	}
	return slices.Contains(graphiQLProxiedPaths, p)
}

type AdminAPIHandler struct {
	Database *globaldb.Handle
	Authz    AdminAPIAuthzService
	AdminAPI AdminAPIService
}

func (h *AdminAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := AdminAPILogger.GetLogger(ctx)
	resolved := relay.FromGlobalID(httproute.GetParam(r, "appid"))
	if resolved == nil || resolved.Type != "App" {
		logger.Debug(ctx, "invalid app ID", slog.Any("resolved", resolved))
		http.Error(w, "invalid app ID", http.StatusBadRequest)
		return
	}

	p := httproute.GetParam(r, "path")

	appID := resolved.ID

	// Every request is access controlled, except for serving the GraphiQL
	// document itself.
	actorUserID := ""
	if !isGraphiQLDocumentRequest(r.Method, p) {
		sessionInfo := session.GetValidSessionInfo(ctx)
		if sessionInfo == nil {
			logger.Debug(ctx, "access to admin API requires authenticated user")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		appIDs, err := h.Authz.ListAuthorizedApps(ctx, sessionInfo.UserID)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to list authorized apps")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		found := slices.Contains(appIDs, appID)
		if !found {
			logger.Debug(ctx, "authenticated user does not have access to the app")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		actorUserID = sessionInfo.UserID
	}

	director, err := h.AdminAPI.Director(ctx, appID, p, actorUserID, service.UsageProxy)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to proxy admin API request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	proxy := httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
}
