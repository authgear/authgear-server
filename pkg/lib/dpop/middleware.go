package dpop

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

type Middleware struct {
	DPoPProvider *Provider
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		dpopHeader := r.Header.Values("DPoP")
		if len(dpopHeader) == 0 {
			next.ServeHTTP(rw, r)
			return
		}
		// https://datatracker.ietf.org/doc/html/rfc9449#name-checking-dpop-proofs
		// Check there is not more than one DPoP HTTP request header field
		if len(dpopHeader) > 1 {
			http.Error(rw, "multiple DPoP headers are not allowed", http.StatusBadRequest)
			return
		}
		dpopJwt := dpopHeader[0]
		proof, err := m.DPoPProvider.ParseProof(dpopJwt)
		if err != nil {
			m.handleError(rw, err)
			return
		}

		if err := m.DPoPProvider.CompareHTM(proof, r.Method); err != nil {
			m.handleError(rw, err)
			return
		}

		if err := m.DPoPProvider.CompareHTU(proof, r); err != nil {
			m.handleError(rw, err)
			return
		}

		r = r.WithContext(WithDPoPProof(r.Context(), proof))
		next.ServeHTTP(rw, r)
	})
}

func (m *Middleware) handleError(rw http.ResponseWriter, err error) {
	var oauthErr *protocol.OAuthProtocolError
	if errors.As(err, &oauthErr) {
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Cache-Control", "no-store")
		rw.Header().Set("Pragma", "no-cache")
		errJson := oauthErr.Response
		errJsonStr, _ := json.Marshal(errJson)
		statusCode := oauthErr.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusBadRequest
		}
		http.Error(rw, string(errJsonStr), statusCode)
		return
	} else {
		panic(err)
	}
}
