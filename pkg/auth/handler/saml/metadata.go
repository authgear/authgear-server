package saml

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureMetadataRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/saml2/metadata/:entity_id_b64")
}

type MetadataHandler struct {
	SAMLConfig *config.SAMLConfig
}

func (h *MetadataHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	entityIDb64 := httproute.GetParam(r, "entity_id_b64")
	entityID, err := saml.DecodeEntityIDURLComponent(entityIDb64)
	if err != nil {
		http.NotFound(rw, r)
		return
	}
	entity, ok := h.SAMLConfig.ResolveProvider(entityID)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	rw.Write([]byte(entity.ID))

}
