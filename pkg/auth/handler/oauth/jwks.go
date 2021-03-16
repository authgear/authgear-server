package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureJWKSRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "OPTIONS").
		WithPathPattern("/oauth2/jwks")
}

type JWSSource interface {
	GetPublicKeySet() (jwk.Set, error)
}

type JWKSHandlerLogger struct{ *log.Logger }

func NewJWKSHandlerLogger(lf *log.Factory) JWKSHandlerLogger {
	return JWKSHandlerLogger{lf.New("handler-jwks")}
}

type JWKSHandler struct {
	Logger JWKSHandlerLogger
	JWKS   JWSSource
}

func (h *JWKSHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	jwks, err := h.JWKS.GetPublicKeySet()
	if err != nil {
		h.Logger.WithError(err).Error("failed to extract public keys")
		http.Error(rw, "internal server error", 500)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	err = encoder.Encode(jwks)
	if err != nil {
		h.Logger.WithError(err).Error("failed to encode public keys")
		http.Error(rw, "internal server error", 500)
		return
	}
}
