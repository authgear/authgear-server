package transport

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
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

type AdminAPIAuthzAdder interface {
	AddAuthz(appID config.AppID, authKey *config.AdminAPIAuthKey, hdr http.Header) (err error)
}

type AdminAPIHandler struct {
	ConfigResolver   AdminAPIConfigResolver
	EndpointResolver AdminAPIEndpointResolver
	AuthzAdder       AdminAPIAuthzAdder
}

func (h *AdminAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resolved := relay.FromGlobalID(httproute.GetParam(r, "appid"))
	if resolved == nil || resolved.Type != "App" {
		http.Error(w, "invalid app ID", http.StatusBadRequest)
		return
	}

	appID := resolved.ID

	cfg, err := h.ConfigResolver.ResolveConfig(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authKey, ok := cfg.SecretConfig.LookupData(config.AdminAPIAuthKeyKey).(*config.AdminAPIAuthKey)
	if !ok {
		http.Error(w, "failed to look up admin API auth key", http.StatusInternalServerError)
		return
	}

	endpoint, err := h.EndpointResolver.ResolveEndpoint(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = endpoint
			req.Host = endpoint.Host
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
