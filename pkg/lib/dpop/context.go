package dpop

import "context"

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
		return nil
	}
	return actx.DPoPProof
}

func GetDPoPProofJKT(ctx context.Context) (string, bool, error) {
	actx := getContext(ctx)
	if actx == nil || actx.DPoPProof == nil {
		return "", false, nil
	}
	proof, err := actx.DPoPProof.Get()
	if err != nil {

		return "", false, err
	}

	return proof.JKT, true, nil

}
