package saml

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureLoginRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/saml2/login/:service_provider_id")
}

type LoginHandlerSAMLService interface {
	IdPMetadata() *saml.Metadata
}

type LoginHandler struct {
	SAMLConfig  *config.SAMLConfig
	SAMLService MetadataHandlerSAMLService
}

func (h *LoginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serviceProviderId := httproute.GetParam(r, "service_provider_id")
	_, ok := h.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	// TODO
	rw.Write([]byte("ok"))
}
