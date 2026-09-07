package dpop

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type ProofReplayStore interface {
	MarkProofUsed(ctx context.Context, proof *DPoPProof) (alreadyUsed bool, err error)
}

type Middleware struct {
	DPoPProvider     *Provider
	ProofReplayStore ProofReplayStore
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
			m.handleError(ctx, next, rw, r, err)
			return
		}

		if err := m.DPoPProvider.CompareHTM(proof, r.Method); err != nil {
			m.handleError(ctx, next, rw, r, err)
			return
		}

		if err := m.DPoPProvider.CompareHTU(proof, r); err != nil {
			m.handleError(ctx, next, rw, r, err)
			return
		}

		// Only once the proof is otherwise valid and bound to this request is
		// it worth remembering, so that a malformed or misdirected proof
		// cannot burn a jti for the legitimate holder of the key.
		// https://datatracker.ietf.org/doc/html/rfc9449#section-11.1
		//
		// The replay check is best effort. If the store cannot be reached the
		// proof has still been parsed, had its signature verified, and been
		// matched against this request's method and URI; only the guarantee
		// that it has not been presented before is lost. Rejecting the request
		// instead would turn a Redis outage into an outage of every
		// DPoP-authenticated endpoint, which is the worse failure.
		alreadyUsed, err := m.ProofReplayStore.MarkProofUsed(ctx, proof)
		switch {
		case err != nil:
			logger := middlewareLogger.GetLogger(ctx)
			logger.WithError(err).Error(ctx,
				"failed to record DPoP proof; replay is not being prevented",
				slog.Bool("dpop_logs", true),
			)
		case alreadyUsed:
			m.handleError(ctx, next, rw, r, ErrProofReplayed)
			return
		}

		r = r.WithContext(WithDPoPProof(ctx, proof))
		next.ServeHTTP(rw, r)
	})
}

func (m *Middleware) handleError(ctx context.Context, next http.Handler, rw http.ResponseWriter, r *http.Request, err error) {
	logger := middlewareLogger.GetLogger(ctx)
	var oauthErr *protocol.OAuthProtocolError
	// If it is an dpop error, we do not return error here.
	// Continue to serve the request, until someone need the proof and handle error there.
	if errors.As(err, &oauthErr) {
		logger.WithError(err).Warn(ctx,
			"failed to parse dpop proof",
			slog.Bool("dpop_logs", true),
		)
		r = r.WithContext(WithDPoPProof(ctx, &InvalidDPoPProofWithError{
			Error: err,
		}))
		next.ServeHTTP(rw, r)
	} else {
		panic(err)
	}
}
