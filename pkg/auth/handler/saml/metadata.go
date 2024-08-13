package saml

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/saml"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureMetadataRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/saml2/metadata/:service_provider_id")
}

type MetadataHandlerSAMLService interface {
	IdpMetadata(serviceProviderId string) *saml.Metadata
}

type MetadataHandler struct {
	SAMLConfig  *config.SAMLConfig
	SAMLService MetadataHandlerSAMLService
}

func (h *MetadataHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serviceProviderId := httproute.GetParam(r, "service_provider_id")
	sp, ok := h.SAMLConfig.ResolveProvider(serviceProviderId)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	metadataBytes := h.SAMLService.IdpMetadata(sp.ID).ToXMLBytes()
	fileName := fmt.Sprintf("%s-metadata.xml", serviceProviderId)
	rw.Header().Set("Content-Type", "application/samlmetadata+xml")
	rw.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	_, err := rw.Write(metadataBytes)
	if err != nil {
		panic(err)
	}
}
