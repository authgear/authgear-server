package dpop

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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
			if apierrors.IsAPIError(err) {
				apierr := apierrors.AsAPIError(err)
				http.Error(rw, fmt.Sprintf("%s:%s", apierr.Reason, apierr.Message), apierr.Name.HTTPStatus())
				return
			} else {
				panic(err)
			}
		}
		WithDPoPProof(r.Context(), proof)

	})
}
