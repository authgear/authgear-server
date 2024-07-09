package dpop

import "context"

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	DPoPJwt *DPoPJwt
}

func getContext(ctx context.Context) *contextValue {
	actx, _ := ctx.Value(contextKey).(*contextValue)
	return actx
}

func WithDPoPJwt(ctx context.Context, jwt *DPoPJwt) context.Context {
	actx := getContext(ctx)
	if actx == nil {
		actx = &contextValue{}
	}
	actx.DPoPJwt = jwt
	return context.WithValue(ctx, contextKey, actx)
}

func GetDPoPJwt(ctx context.Context) *DPoPJwt {
	actx := getContext(ctx)
	if actx == nil || actx.DPoPJwt == nil {
		return nil
	}
	return actx.DPoPJwt
}
