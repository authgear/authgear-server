package transport

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"

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

	// Since we serve GraphiQL with GET, we do not impose access control checking, when the method is GET.
	// The access control checking is done when some query is executed with method POST.
	if r.Method == "GET" {
		emptyActorUserID := ""
		director, err := h.AdminAPI.Director(ctx, appID, p, emptyActorUserID, service.UsageProxy)
		if err != nil {
			logger.WithError(err).Error(ctx, "failed to proxy admin API request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		proxy := httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
		return
	}

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

	found := false
	for _, id := range appIDs {
		if id == appID {
			found = true
			break
		}
	}
	if !found {
		logger.Debug(ctx, "authenticated user does not have access to the app")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	director, err := h.AdminAPI.Director(ctx, appID, p, sessionInfo.UserID, service.UsageProxy)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to proxy admin API request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	proxy := httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
}
