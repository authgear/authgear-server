package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/lestrrat-go/jwx/jwk"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func AttachJWKSHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	handler := pkg.MakeHandler(authDependency, newJWKSHandler)
	router.NewRoute().
		Path("/oauth2/jwks").
		Handler(handler).
		Methods("GET")
}

type JWKSHandler struct {
	config config.OIDCConfiguration
}

func (h *JWKSHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	jwks := jwk.Set{}
	for _, key := range h.config.Keys {
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(key.PublicKey))
		if err != nil {
			http.Error(rw, err.Error(), 500)
			return
		}

		k, err := jwk.New(pubKey)
		if err != nil {
			http.Error(rw, err.Error(), 500)
			return
		}

		k.Set(jwk.KeyUsageKey, jwk.ForSignature)
		k.Set(jwk.AlgorithmKey, "RS256")
		k.Set(jwk.KeyIDKey, key.KID)
		jwks.Keys = append(jwks.Keys, k)
	}

	rw.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(rw)
	err := encoder.Encode(jwks)
	if err != nil {
		http.Error(rw, err.Error(), 500)
	}
}
