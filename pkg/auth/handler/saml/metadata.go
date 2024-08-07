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
		WithPathPattern("/saml2/metadata/:entity_id_b64")
}

type MetadataHandlerSAMLService interface {
	IdPMetadata() *saml.Metadata
}

type MetadataHandler struct {
	SAMLConfig  *config.SAMLConfig
	SAMLService MetadataHandlerSAMLService
}

func (h *MetadataHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	entityIDb64 := httproute.GetParam(r, "entity_id_b64")
	entityID, err := saml.DecodeEntityIDURLComponent(entityIDb64)
	if err != nil {
		http.NotFound(rw, r)
		return
	}
	_, ok := h.SAMLConfig.ResolveProvider(entityID)
	if !ok {
		http.NotFound(rw, r)
		return
	}

	metadataBytes := h.SAMLService.IdPMetadata().ToXMLBytes()
	fileName := fmt.Sprintf("%s-metadata.xml", entityIDb64)
	rw.Header().Set("Content-Type", "application/samlmetadata+xml")
	rw.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	_, err = rw.Write(metadataBytes)
	if err != nil {
		panic(err)
	}
}
