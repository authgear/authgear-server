package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/httproute"
)

func ConfigureOIDCMetadataRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "OPTIONS").
		WithPathPattern("/.well-known/openid-configuration")
}

func ConfigureOAuthMetadataRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "OPTIONS").
		WithPathPattern("/.well-known/oauth-authorization-server")
}

type MetadataProvider interface {
	PopulateMetadata(meta map[string]interface{})
}

type MetadataHandler struct {
	Providers []MetadataProvider
}

func (h *MetadataHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	meta := map[string]interface{}{}
	for _, provider := range h.Providers {
		provider.PopulateMetadata(meta)
	}

	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	err := encoder.Encode(meta)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
}
