package dpop

import "context"

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	DPoPProof *DPoPProof
}

func getContext(ctx context.Context) *contextValue {
	actx, _ := ctx.Value(contextKey).(*contextValue)
	return actx
}

func WithDPoPProof(ctx context.Context, proof *DPoPProof) context.Context {
	actx := getContext(ctx)
	if actx == nil {
		actx = &contextValue{}
	}
	actx.DPoPProof = proof
	return context.WithValue(ctx, contextKey, actx)
}

func GetDPoPProof(ctx context.Context) *DPoPProof {
	actx := getContext(ctx)
	if actx == nil || actx.DPoPProof == nil {
		return nil
	}
	return actx.DPoPProof
}

func GetDPoPProofJKT(ctx context.Context) (string, bool) {
	actx := getContext(ctx)
	if actx == nil || actx.DPoPProof == nil {
		return "", false
	}
	return actx.DPoPProof.JKT, true
}
