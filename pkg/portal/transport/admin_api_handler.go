package transport

import (
	"net/http"
	"net/http/httputil"

	"github.com/authgear/graphql-go-relay"

	"github.com/authgear/authgear-server/pkg/portal/db"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureAdminAPIRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET", "POST").WithPathPattern("/api/apps/:appid/graphql")
}

type AdminAPIAuthzService interface {
	ListAuthorizedApps(userID string) ([]string, error)
}

type AdminAPIService interface {
	Director(appID string) (func(*http.Request), error)
}

type AdminAPILogger struct{ *log.Logger }

func NewAdminAPILogger(lf *log.Factory) AdminAPILogger {
	return AdminAPILogger{lf.New("admin-api-proxy")}
}

type AdminAPIHandler struct {
	Database *db.Handle
	Authz    AdminAPIAuthzService
	AdminAPI AdminAPIService
	Logger   AdminAPILogger
}

func (h *AdminAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resolved := relay.FromGlobalID(httproute.GetParam(r, "appid"))
	if resolved == nil || resolved.Type != "App" {
		h.Logger.Debugf("invalid app ID: %v", resolved)
		http.Error(w, "invalid app ID", http.StatusBadRequest)
		return
	}

	appID := resolved.ID

	sessionInfo := session.GetValidSessionInfo(r.Context())
	if sessionInfo == nil {
		h.Logger.Debugf("access to admin API requires authenticated user")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var appIDs []string
	err := h.Database.ReadOnly(func() (err error) {
		appIDs, err = h.Authz.ListAuthorizedApps(sessionInfo.UserID)
		return
	})
	if err != nil {
		h.Logger.WithError(err).Errorf("failed to list authorized apps")
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
		h.Logger.Debugf("authenticated user does not have access to the app")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	director, err := h.AdminAPI.Director(appID)
	if err != nil {
		h.Logger.WithError(err).Errorf("failed to proxy admin API request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	proxy := httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
}
