package transport

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/authgear/graphql-go-relay"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureAdminAPIRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET", "POST").WithPathPattern("/api/apps/:appid/graphql")
}

type AdminAPIConfigResolver interface {
	ResolveConfig(appID string) (*config.Config, error)
}

type AdminAPIEndpointResolver interface {
	ResolveEndpoint(appID string) (url *url.URL, err error)
}

type AdminAPIHostResolver interface {
	ResolveHost(appID string) (host string, err error)
}

type AdminAPIAuthzAdder interface {
	AddAuthz(appID config.AppID, authKey *config.AdminAPIAuthKey, hdr http.Header) (err error)
}

type AdminAPIAuthzService interface {
	ListAuthorizedApps(userID string) ([]string, error)
}

type AdminAPILogger struct{ *log.Logger }

func NewAdminAPILogger(lf *log.Factory) AdminAPILogger {
	return AdminAPILogger{lf.New("admin-api-proxy")}
}

type AdminAPIHandler struct {
	Authz            AdminAPIAuthzService
	ConfigResolver   AdminAPIConfigResolver
	EndpointResolver AdminAPIEndpointResolver
	HostResolver     AdminAPIHostResolver
	AuthzAdder       AdminAPIAuthzAdder
	Logger           AdminAPILogger
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

	appIDs, err := h.Authz.ListAuthorizedApps(sessionInfo.UserID)
	if err != nil {
		h.Logger.WithError(err).Debugf("failed to list authorized apps")
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

	cfg, err := h.ConfigResolver.ResolveConfig(appID)
	if err != nil {
		h.Logger.WithError(err).Debugf("failed to resolve config: %v", appID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authKey, ok := cfg.SecretConfig.LookupData(config.AdminAPIAuthKeyKey).(*config.AdminAPIAuthKey)
	if !ok {
		h.Logger.Debugf("failed to look up admin API auth key: %v", appID)
		http.Error(w, "failed to look up admin API auth key", http.StatusInternalServerError)
		return
	}

	endpoint, err := h.EndpointResolver.ResolveEndpoint(appID)
	if err != nil {
		h.Logger.WithError(err).Debugf("failed to resolve endpoint: %v", appID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	host, err := h.HostResolver.ResolveHost(appID)
	if err != nil {
		h.Logger.WithError(err).Debugf("failed to resolve host: %v", appID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.WithFields(map[string]interface{}{
		"appID":    appID,
		"endpoint": endpoint.String(),
		"host":     host,
	}).Debugf("proxy admin API")

	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = endpoint

			// We have to set both to ensure Admin API server sees the correct host.
			req.Host = host
			req.Header.Set("X-Forwarded-Host", req.Host)

			err = h.AuthzAdder.AddAuthz(
				config.AppID(appID),
				authKey,
				req.Header,
			)
			if err != nil {
				panic(err)
			}
		},
	}

	proxy.ServeHTTP(w, r)
}
