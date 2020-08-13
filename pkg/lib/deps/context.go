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

func getRequestProvider(r *http.Request) *RequestProvider {
	pCtx := r.Context().Value(providerContextKey).(*providerContext)
	p := pCtx.p.NewRequestProvider(r)
	return p
}

func getTaskProvider(ctx context.Context) *TaskProvider {
	pCtx := ctx.Value(providerContextKey).(*providerContext)
	p := pCtx.p.NewTaskProvider(ctx)
	return p
}
