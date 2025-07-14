package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

func ConfigureJWKSRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "OPTIONS").
		WithPathPattern("/oauth2/jwks")
}

type JWSSource interface {
	GetPublicKeySet() (jwk.Set, error)
}

var JWKSHandlerLogger = slogutil.NewLogger("handler-jwks")

type JWKSHandler struct {
	JWKS JWSSource
}

func (h *JWKSHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := JWKSHandlerLogger.GetLogger(ctx)
	jwks, err := h.JWKS.GetPublicKeySet()
	if err != nil {
		logger.WithError(err).Error(r.Context(), "failed to extract public keys")
		http.Error(rw, "internal server error", 500)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(jwks)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to encode public keys")
		http.Error(rw, "internal server error", 500)
		return
	}
}
