package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
)

func AttachMetadataHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	handler := pkg.MakeHandler(authDependency, newMetadataHandler)
	router.NewRoute().
		Path("/.well-known/openid-configuration").
		Handler(handler).
		Methods("GET", "OPTIONS")
	router.NewRoute().
		Path("/.well-known/oauth-authorization-server").
		Handler(handler).
		Methods("GET", "OPTIONS")
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
