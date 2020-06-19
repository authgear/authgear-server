package deps

import "context"

type providerContextKeyType struct{}

var providerContextKey = providerContextKeyType{}

func withProvider(ctx context.Context, p interface{}) context.Context {
	return context.WithValue(ctx, providerContextKey, p)
}

func getRequestProvider(ctx context.Context) *RequestProvider {
	p := ctx.Value(providerContextKey).(*RequestProvider)
	return p
}

func getTaskProvider(ctx context.Context) *TaskProvider {
	p := ctx.Value(providerContextKey).(*TaskProvider)
	return p
}
