package dpop

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	DPoPProof MaybeDPoPProof
}

func getContext(ctx context.Context) *contextValue {
	actx, _ := ctx.Value(contextKey).(*contextValue)
	return actx
}

func WithDPoPProof(ctx context.Context, proof MaybeDPoPProof) context.Context {
	actx := getContext(ctx)
	if actx == nil {
		actx = &contextValue{}
	}
	actx.DPoPProof = proof
	return context.WithValue(ctx, contextKey, actx)
}

func GetDPoPProof(ctx context.Context) MaybeDPoPProof {
	actx := getContext(ctx)
	if actx == nil || actx.DPoPProof == nil {
		return &MissingDPoPProof{}
	}
	return actx.DPoPProof
}

func GetDPoPProofJKT(ctx context.Context, client *config.OAuthClientConfig) (string, bool, error) {
	actx := getContext(ctx)
	if actx == nil || actx.DPoPProof == nil {
		return "", false, nil
	}
	proof, err := actx.DPoPProof.Get()
	if err != nil {
		if !client.DPoPDisabled {
			return "", false, err
		} else {
			return "", false, nil
		}
	}

	return proof.JKT, true, nil

}
