package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func ConfigureMetadataHandler(router *mux.Router, h http.Handler) {
	router.NewRoute().
		Path("/.well-known/openid-configuration").
		Methods("GET", "OPTIONS").
		Handler(h)
	router.NewRoute().
		Path("/.well-known/oauth-authorization-server").
		Methods("GET", "OPTIONS").
		Handler(h)
}

type oauthMetadataProvider interface {
	PopulateMetadata(meta map[string]interface{})
}

type MetadataHandler struct {
	metaProviders []oauthMetadataProvider
}

func (h *MetadataHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	meta := map[string]interface{}{}
	for _, provider := range h.metaProviders {
		provider.PopulateMetadata(meta)
	}

	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	err := encoder.Encode(meta)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
}
