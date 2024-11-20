package deps

import (
	"context"
	"net/http"
)

type providerContextKeyType struct{}

var providerContextKey = providerContextKeyType{}

type providerContext struct {
	p *AppProvider
}

func withProvider(ctx context.Context, p *AppProvider) context.Context {
	return context.WithValue(ctx, providerContextKey, &providerContext{
		p: p,
	})
}

func getRequestProvider(w http.ResponseWriter, r *http.Request) *RequestProvider {
	pCtx := r.Context().Value(providerContextKey).(*providerContext)
	p := pCtx.p.NewRequestProvider(w, r)
	return p
}
