package saml

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/saml/samlprotocol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureMetadataRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET").
		WithPathPattern("/saml2/metadata/:service_provider_id")
}

type MetadataHandler struct {
	SAMLService HandlerSAMLService
}

func (h *MetadataHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	serviceProviderId := httproute.GetParam(r, "service_provider_id")

	metadata, err := h.SAMLService.IdpMetadata(serviceProviderId)
	if err != nil {
		if errors.Is(err, samlprotocol.ErrServiceProviderNotFound) {
			http.NotFound(rw, r)
			return
		}
		panic(err)
	}

	metadataBytes := metadata.ToXMLBytes()
	fileName := fmt.Sprintf("%s-metadata.xml", serviceProviderId)
	rw.Header().Set("Content-Type", "application/samlmetadata+xml")
	rw.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	_, err = rw.Write(metadataBytes)
	if err != nil {
		panic(err)
	}
}
