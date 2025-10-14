package dpop

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type Middleware struct {
	DPoPProvider *Provider
}

var middlewareLogger = slogutil.NewLogger("dpop-middleware")

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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
			m.handleError(ctx, rw, err)
			return
		}

		if err := m.DPoPProvider.CompareHTM(proof, r.Method); err != nil {
			m.handleError(ctx, rw, err)
			return
		}

		if err := m.DPoPProvider.CompareHTU(proof, r); err != nil {
			m.handleError(ctx, rw, err)
			return
		}

		r = r.WithContext(WithDPoPProof(r.Context(), proof))
		next.ServeHTTP(rw, r)
	})
}

func (m *Middleware) handleError(ctx context.Context, rw http.ResponseWriter, err error) {
	logger := middlewareLogger.GetLogger(ctx)
	var oauthErr *protocol.OAuthProtocolError
	if errors.As(err, &oauthErr) {
		logger.WithSkipLogging().WithError(oauthErr).Error(ctx,
			"failed to parse dpop proof",
			slog.Bool("dpop_logs", true),
		)
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Cache-Control", "no-store")
		rw.Header().Set("Pragma", "no-cache")
		errJson := oauthErr.Response
		errJsonStr, _ := json.Marshal(errJson)
		statusCode := oauthErr.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusBadRequest
		}
		rw.WriteHeader(statusCode)
		_, err := rw.Write([]byte(errJsonStr))
		if err != nil {
			panic(err)
		}
		return
	} else {
		panic(err)
	}
}
